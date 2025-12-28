package dto

import (
	z "github.com/Oudwins/zog"
	"go.lumeweb.com/httputil"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
)

var _ httputil.DTORequest[*models.Item] = (*UpdateItemRequest)(nil)
var _ httputil.DTOResponse[*models.Item] = (*ItemResponse)(nil)
var _ httputil.DTOValidator = (*UpdateItemRequest)(nil)

// UpdateItemRequest represents the request for updating an existing item.
type UpdateItemRequest struct {
	ID          uint   `param:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Schema returns the zog validation schema for UpdateItemRequest.
func (r *UpdateItemRequest) Schema() *z.StructSchema {
	return z.Struct(z.Shape{
		"ID":          z.Uint().GT(0),
		"Name":        z.String().Required().Min(1).Max(255),
		"Description": z.String().Max(1000).Optional(),
	})
}

// ToModel converts the DTO to a model.
func (r *UpdateItemRequest) ToModel() (*models.Item, error) {
	return &models.Item{
		Name:        r.Name,
		Description: r.Description,
	}, nil
}
