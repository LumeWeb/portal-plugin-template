package dto

import (
	"time"

	z "github.com/Oudwins/zog"
	"go.lumeweb.com/httputil"
	"go.lumeweb.com/portal-plugin-template/internal/db/models"
)

var _ httputil.DTORequest[*models.Item] = (*CreateItemRequest)(nil)
var _ httputil.DTOResponse[*models.Item] = (*ItemResponse)(nil)
var _ httputil.DTOValidator = (*CreateItemRequest)(nil)

// CreateItemRequest represents the request body for creating a new item.
type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Schema returns the zog validation schema for CreateItemRequest.
func (r *CreateItemRequest) Schema() *z.StructSchema {
	return z.Struct(z.Shape{
		"Name":        z.String().Required().Min(1).Max(255),
		"Description": z.String().Max(1000).Optional(),
	})
}

// ToModel converts the DTO to a model.
func (r *CreateItemRequest) ToModel() (*models.Item, error) {
	return &models.Item{
		Name:        r.Name,
		Description: r.Description,
	}, nil
}

// ItemResponse represents the response for item operations.
type ItemResponse struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// FromModel populates the DTO from a model.
func (r *ItemResponse) FromModel(item *models.Item) error {
	r.ID = item.ID
	r.Name = item.Name
	r.Description = item.Description
	r.CreatedAt = item.CreatedAt
	r.UpdatedAt = item.UpdatedAt
	return nil
}
