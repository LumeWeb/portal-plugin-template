package protocol

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	pluginCore "go.lumeweb.com/portal-plugin-template/core"
	"go.lumeweb.com/portal-plugin-template/internal"
	protocolConfig "go.lumeweb.com/portal-plugin-template/internal/config"
	"go.lumeweb.com/portal-plugin-template/internal/protocol/request"
	"go.lumeweb.com/portal/config"
	"go.lumeweb.com/portal/core"
	"go.lumeweb.com/portal/db/models"
	"go.lumeweb.com/portal/event"
	"go.lumeweb.com/portal/service"
	"go.uber.org/zap"
)

var _ core.Protocol = (*Protocol)(nil)

// Protocol handles storage operations for the plugin.
type Protocol struct {
	*core.BaseComponent
	itemService core.Service
	storage     core.StorageService
	coordinator core.WorkflowCoordinator

	uploads   map[string]*uploadState
	uploadsMu sync.RWMutex
	isRunning bool
}

type uploadState struct {
	ID        string
	Size      uint64
	Uploaded  uint64
	Started   time.Time
	Completed bool
	Hash      string
}

func (p *Protocol) ID() string {
	return internal.PLUGIN_NAME
}

func (p *Protocol) Name() string {
	return internal.PLUGIN_NAME
}

func (p *Protocol) DisplayName() string {
	return "Template Plugin"
}

func (p *Protocol) Operations() []core.Operation {
	return []core.Operation{
		core.NewStoreOperation(p.Name(), &storeHandler{protocol: p}),
		core.NewOperation(
			fmt.Sprintf("%s.scan", p.Name()),
			core.OpTypeScan,
			&scanHandler{protocol: p},
		),
	}
}

func (p *Protocol) Workflows() []core.WorkflowDefinition {
	return nil
}

func (p *Protocol) GetConfig() config.ProtocolConfig {
	return &protocolConfig.ProtocolConfig{}
}

func NewProtocol() (core.Protocol, []core.ContextBuilderOption, error) {
	proto := &Protocol{
		uploads:   make(map[string]*uploadState),
		uploadsMu: sync.RWMutex{},
	}

	opts := core.ContextOptions(
		core.ContextWithStartupFunc(func(ctx core.Context) error {
			proto.storage = core.GetService[core.StorageService](ctx, core.STORAGE_SERVICE)
			proto.coordinator = core.GetService[core.WorkflowService](ctx, core.WORKFLOW_SERVICE)
			proto.itemService = core.GetService[pluginCore.ItemService](ctx, pluginCore.ITEM_SERVICE)
			requestSvc := ctx.Service(core.REQUEST_SERVICE).(core.RequestService)
			requestSvc.RegisterRequestModel(proto.Name(), &request.TemplateRequest{})

			event.OnBootCompleted(ctx, func(c core.Context, ctx context.Context) error {
				cfg := core.GetProtocolConfig[*protocolConfig.ProtocolConfig](c, internal.PLUGIN_NAME)

				proto.Logger().Info("Template protocol initialized",
					zap.String("storage_path", cfg.StoragePath),
					zap.Int("max_items", cfg.MaxItems),
					zap.Bool("cache_enabled", cfg.CacheEnabled))

				return nil
			})

			return nil
		}),
	)

	return proto, opts, nil
}

func (p *Protocol) Start(_ core.Context) error {
	p.Logger().Info("Starting template protocol")
	p.isRunning = true
	return nil
}

func (p *Protocol) Stop(_ core.Context) error {
	p.Logger().Info("Stopping template protocol")
	p.isRunning = false
	return nil
}

func (p *Protocol) EncodeFileName(hash core.StorageHash) string {
	return hash.Multihash().B58String()
}

func (p *Protocol) Hash(r io.Reader, _ uint64) (core.StorageHash, error) {
	h := sha256.New()
	if _, err := io.Copy(h, r); err != nil {
		return nil, err
	}
	return core.NewStorageHash(h.Sum(nil), uint64(sha256.Size), 0, nil), nil
}

func (p *Protocol) HandleUpload(ctx context.Context, reader io.Reader, size uint64) (core.StorageHash, error) {
	if !p.isRunning {
		return nil, fmt.Errorf("protocol not running")
	}

	h := sha256.New()
	if _, err := io.Copy(h, reader); err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	hash := core.NewStorageHash(h.Sum(nil), uint64(sha256.Size), 0, nil)

	req := &models.Request{
		Protocol: p.Name(),
		Hash:     hash.Multihash(),
	}

	_, err := p.coordinator.StartWorkflow(ctx, "template.upload", core.WithWorkflowRequestData(req))
	if err != nil {
		return nil, fmt.Errorf("failed to start upload workflow: %w", err)
	}

	state := &uploadState{
		ID:      fmt.Sprintf("%d", req.ID),
		Size:    size,
		Started: time.Now(),
		Hash:    hash.Multihash().B58String(),
	}

	p.uploadsMu.Lock()
	p.uploads[state.ID] = state
	p.uploadsMu.Unlock()

	return hash, nil
}

func (p *Protocol) GetUploadStatus(uploadID string) (*uploadState, error) {
	requestID, err := strconv.ParseUint(uploadID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid upload ID: %w", err)
	}

	status, err := p.coordinator.GetWorkflowStatus(context.Background(), uint(requestID))
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow status: %w", err)
	}

	return &uploadState{
		ID:        uploadID,
		Size:      0,
		Started:   status.StartedAt,
		Completed: status.Status == models.RequestStatusCompleted,
		Hash:      "",
	}, nil
}

// storeHandler handles the store operation.
type storeHandler struct {
	protocol *Protocol
}

func (h *storeHandler) ValidateRequest(_ context.Context, _ *models.Request) error {
	return nil
}

func (h *storeHandler) Execute(ctx context.Context, req *models.Request) error {
	logger := h.protocol.Logger()

	storage := h.protocol.BaseComponent.Context().Service(core.STORAGE_SERVICE).(core.StorageService)
	storageProtocol := h.protocol

	readCloser, err := storage.S3GetTemporaryUpload(ctx, storageProtocol, fmt.Sprintf("%d", req.ID))
	if err != nil {
		return fmt.Errorf("failed to get temporary upload: %w", err)
	}
	defer readCloser.Close()

	data, err := io.ReadAll(readCloser)
	if err != nil {
		return fmt.Errorf("failed to read upload: %w", err)
	}

	uploadReq := service.NewStorageUploadRequest(
		core.StorageUploadWithProtocol(storageProtocol),
		core.StorageUploadWithData(bytes.NewReader(data)),
		core.StorageUploadWithSize(uint64(len(data))),
		core.StorageUploadWithProof(core.NewStorageHashFromMultihashBytes(req.Hash, 0, nil)),
	)

	_, err = storage.UploadObject(ctx, uploadReq)
	if err != nil {
		return fmt.Errorf("failed to store object: %w", err)
	}

	if err := storage.S3DeleteTemporaryUpload(ctx, storageProtocol, fmt.Sprintf("%d", req.ID)); err != nil {
		logger.Error("failed to cleanup temporary upload", zap.Error(err))
	}

	return nil
}

func (h *storeHandler) GetStatus(_ context.Context, _ *models.Request) (*core.RequestStatus, error) {
	return &core.RequestStatus{State: "completed", Message: "Upload completed"}, nil
}

func (h *storeHandler) Cleanup(_ context.Context, _ *models.Request) error {
	return nil
}

// scanHandler handles the scan operation.
type scanHandler struct {
	protocol *Protocol
}

func (h *scanHandler) ValidateRequest(_ context.Context, _ *models.Request) error {
	return nil
}

func (h *scanHandler) Execute(_ context.Context, _ *models.Request) error {
	return nil
}

func (h *scanHandler) GetStatus(_ context.Context, _ *models.Request) (*core.RequestStatus, error) {
	return &core.RequestStatus{State: "completed", Message: "Scan completed"}, nil
}

func (h *scanHandler) Cleanup(_ context.Context, _ *models.Request) error {
	return nil
}
