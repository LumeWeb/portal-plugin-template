package protocol_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"strings"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	pluginCore "go.lumeweb.com/portal-plugin-template/core"
	"go.lumeweb.com/portal-plugin-template/internal"
	protocolConfig "go.lumeweb.com/portal-plugin-template/internal/config"
	"go.lumeweb.com/portal-plugin-template/internal/protocol"
	"go.lumeweb.com/portal-plugin-template/internal/protocol/request"
	"go.lumeweb.com/portal-plugin-template/internal/service"
	"go.lumeweb.com/portal/core"
	coreTesting "go.lumeweb.com/portal/core/testing"
	"go.lumeweb.com/portal/db/models"
)

func TestMain(m *testing.M) {
	coreTesting.WithOptions(m,
		coreTesting.WithServiceFactory(pluginCore.ITEM_SERVICE, service.NewItemService),
		coreTesting.WithProtocol(internal.PLUGIN_NAME, protocol.NewProtocol),
		coreTesting.WithProtocolConfig(internal.PLUGIN_NAME, &protocolConfig.ProtocolConfig{}),
		coreTesting.WithTestMainContextSimple(func(ctx coreTesting.TestContext) []coreTesting.TestContextBuilderOption {
			// Import required for mock.AnythingOfType
			_ = request.TemplateRequest{}
			mockRequestSvc := coreTesting.GetMockRequestService(ctx)
			mockRequestSvc.EXPECT().RegisterRequestModel(internal.PLUGIN_NAME, mock.AnythingOfType("*request.TemplateRequest")).Return()
			return nil
		}),
	)
}

func TestProtocol_StartStop(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get the protocol from the test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to our protocol type to access Start/Stop methods
		templateProto := proto.(*protocol.Protocol)

		// Test Start
		err := templateProto.Start(ctx)
		assert.NoError(t, err)

		// Test Stop
		err = templateProto.Stop(ctx)
		assert.NoError(t, err)
	})
}

func TestProtocol_Hash(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get the protocol from the test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to our protocol type to access Hash method
		templateProto := proto.(*protocol.Protocol)

		testData := []byte("test data for hashing")
		reader := bytes.NewReader(testData)

		hash, err := templateProto.Hash(reader, uint64(len(testData)))
		require.NoError(t, err)
		assert.NotNil(t, hash)

		// Verify hash was created successfully
		assert.NotNil(t, hash)
		assert.NotEmpty(t, hash.Multihash())
	})
}

func TestProtocol_EncodeFileName(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get the protocol from the test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to concrete type to access EncodeFileName method
		templateProto := proto.(*protocol.Protocol)

		// Create a test hash
		testData := []byte("test")
		h := sha256.New()
		h.Write(testData)
		hash := core.NewStorageHash(h.Sum(nil), uint64(sha256.Size), 0, nil)

		// Test encoding
		encoded := templateProto.EncodeFileName(hash)
		assert.NotEmpty(t, encoded)
		// Base58 encoded multihash (format varies, not always Qm prefix)
	})
}

func TestProtocol_HandleUpload(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get the protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to concrete type to access Start and HandleUpload methods
		templateProto := proto.(*protocol.Protocol)

		// Start the protocol
		err := templateProto.Start(ctx)
		require.NoError(t, err)

		// Mock workflow coordinator
		mockWorkflow := coreTesting.GetMockWorkflowService(ctx)
		mockWorkflow.EXPECT().StartWorkflow(mock.Anything, "template.upload", mock.Anything).
			Return(&models.Request{Model: gorm.Model{ID: 1}}, nil)
		mockWorkflow.EXPECT().GetWorkflowStatus(mock.Anything, uint(1)).
			Return(&core.WorkflowStatus{
				Status:    models.RequestStatusCompleted,
				StartedAt: time.Now(),
			}, nil)

		// Test upload
		testData := []byte("test upload data")
		reader := bytes.NewReader(testData)

		hash, err := templateProto.HandleUpload(context.Background(), reader, uint64(len(testData)))
		require.NoError(t, err)
		assert.NotNil(t, hash)

		// Verify upload state was tracked
		status, err := templateProto.GetUploadStatus("1")
		require.NoError(t, err)
		assert.Equal(t, "1", status.ID)
		assert.True(t, status.Completed)

		mockWorkflow.AssertExpectations(t)
	})
}

func TestProtocol_HandleUpload_NotRunning(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to concrete type to access HandleUpload method
		templateProto := proto.(*protocol.Protocol)

		// Don't start the protocol - should fail
		testData := []byte("test data")
		reader := bytes.NewReader(testData)

		_, err := templateProto.HandleUpload(context.Background(), reader, uint64(len(testData)))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not running")
	})
}

func TestProtocol_GetUploadStatus(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to concrete type to access GetUploadStatus method
		templateProto := proto.(*protocol.Protocol)

		// Mock workflow coordinator
		mockWorkflow := coreTesting.GetMockWorkflowService(ctx)
		mockWorkflow.EXPECT().GetWorkflowStatus(mock.Anything, uint(1)).
			Return(&core.WorkflowStatus{
				Status: models.RequestStatusCompleted,
			}, nil)

		// Test getting status
		status, err := templateProto.GetUploadStatus("1")
		require.NoError(t, err)
		assert.Equal(t, "1", status.ID)
		assert.True(t, status.Completed)

		mockWorkflow.AssertExpectations(t)
	})
}

func TestProtocol_GetUploadStatus_InvalidID(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to concrete type to access GetUploadStatus method
		templateProto := proto.(*protocol.Protocol)

		// Test with invalid ID
		_, err := templateProto.GetUploadStatus("invalid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid upload ID")
	})
}

func TestProtocol_Operations(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Test operations are registered
		ops := proto.Operations()
		assert.Len(t, ops, 2)

		// Verify operation names
		opNames := make([]string, len(ops))
		for i, op := range ops {
			opNames[i] = op.Name()
		}
		assert.Contains(t, opNames, "Store")
		assert.Contains(t, opNames, "Scan")
	})
}

func TestProtocol_Workflows(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Test workflows (should be empty for template)
		workflows := proto.Workflows()
		assert.Nil(t, workflows)
	})
}

func TestProtocol_GetConfig(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Test config retrieval
		config := proto.GetConfig()
		assert.NotNil(t, config)
		assert.IsType(t, &protocolConfig.ProtocolConfig{}, config)
	})
}

func TestProtocol_Properties(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Test protocol properties
		assert.Equal(t, internal.PLUGIN_NAME, proto.ID())
		assert.Equal(t, internal.PLUGIN_NAME, proto.Name())
		assert.Equal(t, "Template Plugin", proto.DisplayName())
	})
}

func TestProtocol_Hash_LargeData(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to concrete type to access Hash method
		templateProto := proto.(*protocol.Protocol)

		// Test with larger data
		largeData := strings.Repeat("test data ", 10000)
		reader := strings.NewReader(largeData)

		hash, err := templateProto.Hash(reader, uint64(len(largeData)))
		require.NoError(t, err)
		assert.NotNil(t, hash)

		// Verify hash was created successfully
		assert.NotNil(t, hash)
		assert.NotEmpty(t, hash.Multihash())
	})
}

func TestProtocol_Hash_ReadError(t *testing.T) {
	coreTesting.RunTestCase(t, func(tb coreTesting.TB, ctx coreTesting.TestContext) {
		// Get protocol from test context
		proto := core.GetProtocol(internal.PLUGIN_NAME)
		require.NotNil(t, proto)

		// Cast to concrete type to access Hash method
		templateProto := proto.(*protocol.Protocol)

		// Create a reader that fails
		errorReader := &errorReader{err: io.ErrUnexpectedEOF}

		_, err := templateProto.Hash(errorReader, 100)
		assert.Error(t, err)
	})
}

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}
