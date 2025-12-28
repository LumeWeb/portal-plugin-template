package dto

import "go.lumeweb.com/httputil"

// StatusResponse represents a simple status response for operations.
type StatusResponse struct {
	Status string `json:"status"`
}

// FromModel is a no-op since there's no model to populate from.
// The response is set directly before calling EncodeResponse.
func (r *StatusResponse) FromModel(model *StatusResponse) error {
	return nil
}

// Compile-time interface verification (self-referencing pattern for no-model responses)
var _ httputil.DTOResponse[*StatusResponse] = (*StatusResponse)(nil)
