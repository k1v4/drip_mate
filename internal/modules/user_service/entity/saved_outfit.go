package entity

import "github.com/google/uuid"

type SaveOutfitRequest struct {
	Name           string      `json:"name"             validate:"required,min=1,max=255"`
	CatalogItemIDs []uuid.UUID `json:"catalog_item_ids" validate:"required,min=1,dive,required"`
	LogID          int         `json:"log_id"           validate:"omitempty"`
}

type OutfitItem struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	ImageURL string    `json:"image"`
	Material string    `json:"material"`
}

type Outfit struct {
	ID    uuid.UUID    `json:"id" db:"id"`
	Name  string       `json:"name" db:"name"`
	Items []OutfitItem `json:"items" db:"items"`
}

type UpdateContext struct {
	ID     uuid.UUID `json:"id"`
	City   string    `json:"city"`
	Styles *[]int    `json:"styles"`
	Colors *[]int    `json:"colors"`
	Music  *[]int    `json:"music"`
}
