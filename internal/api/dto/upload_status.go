package dto

// UploadStatusRequest represents the request for checking upload status.
type UploadStatusRequest struct {
	ID string `param:"id" validate:"required,min=1"`
}

// ToModel converts the DTO to a model.
func (r *UploadStatusRequest) ToModel() (*UploadStatusRequest, error) {
	return r, nil
}

// UploadStatusResponse represents the status of an upload operation.
type UploadStatusResponse struct {
	ID        string `json:"id"`
	Size      uint64 `json:"size"`
	Uploaded  uint64 `json:"uploaded"`
	Completed bool   `json:"completed"`
	Hash      string `json:"hash"`
}

// FromModel populates the DTO from an upload state.
func (r *UploadStatusResponse) FromModel(state *UploadState) error {
	r.ID = state.ID
	r.Size = state.Size
	r.Uploaded = state.Uploaded
	r.Completed = state.Completed
	r.Hash = state.Hash
	return nil
}

// UploadState represents the internal upload state.
type UploadState struct {
	ID        string
	Size      uint64
	Uploaded  uint64
	Completed bool
	Hash      string
}
