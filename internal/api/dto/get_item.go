package dto

import (
	z "github.com/Oudwins/zog"
	"go.lumeweb.com/httputil"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
)

var _ httputil.DTORequest[*models.Item] = (*GetItemRequest)(nil)
var _ httputil.DTOValidator = (*GetItemRequest)(nil)

// GetItemRequest represents the request for getting an item by ID.
type GetItemRequest struct {
	ID uint `param:"id"`
}

// Schema returns the zog validation schema for GetItemRequest.
func (r *GetItemRequest) Schema() *z.StructSchema {
	return z.Struct(z.Shape{
		"ID": z.Uint().GT(0),
	})
}

// ToModel converts the DTO to a model.
func (r *GetItemRequest) ToModel() (*models.Item, error) {
	return nil, nil
}
