package entity

import (
	"time"

	"github.com/google/uuid"
)

type Catalog struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Name           string    `db:"name" json:"name" validate:"required"`
	CategoryID     int       `db:"category_id" json:"category_id" validate:"required"`
	Category       string    `db:"category" json:"category"`
	Gender         *string   `db:"gender" json:"gender" validate:"required"`
	SeasonID       int       `db:"season_id" json:"season_id" validate:"required"`
	Season         string    `db:"season" json:"season"`
	FormalityLevel *int16    `db:"formality_level" json:"formality_level" validate:"required"`
	Material       *string   `db:"material" json:"material" validate:"required"`
	ImageURL       string    `db:"image_url" json:"image_url" validate:"required"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
	IsDeleted      bool      `db:"is_deleted" json:"is_deleted"`

	Colors []string `db:"-" json:"colors"`
	Styles []string `db:"-" json:"styles"`
}

type RecommendationsCatalogRequest struct {
	Catalog []Catalog `json:"catalog"`
	LogID   int       `json:"log_id"`
}

type CreateCatalogRequest struct {
	Name           string  `json:"name" form:"name"            validate:"required"`
	CategoryID     int     `json:"category_id" form:"category_id"     validate:"required"`
	Gender         *string `json:"gender" form:"gender"          validate:"required"`
	SeasonID       int     `json:"season_id" form:"season_id"       validate:"required"`
	FormalityLevel *int16  `json:"formality_level" form:"formality_level" validate:"required"`
	Material       *string `json:"material" form:"material"        validate:"required"`
	ImageURL       string  `json:"image_url" form:"image_url"       validate:"omitempty"`
	ColorIDs       []int   `json:"color_ids" form:"color_ids"       validate:"omitempty,dive,min=1"`
	StyleIDs       []int   `json:"style_ids" form:"style_ids"       validate:"omitempty,dive,min=1"`
}

type UpdateCatalogRequest struct {
	ID             uuid.UUID `json:"id" form:"id"`
	Name           string    `json:"name" form:"name"            validate:"omitempty,min=1,max=255"`
	CategoryID     int       `json:"category_id" form:"category_id"     validate:"omitempty,min=1"`
	Gender         *string   `json:"gender" form:"gender"          validate:"omitempty"`
	SeasonID       int       `json:"season_id" form:"season_id"       validate:"omitempty,min=1"`
	FormalityLevel *int16    `json:"formality_level" form:"formality_level" validate:"omitempty"`
	Material       *string   `json:"material" form:"material"        validate:"omitempty"`
	ImageURL       string    `json:"image_url" form:"image_url"       validate:"omitempty,url"`
	ColorIDs       []int     `json:"color_ids" form:"color_ids"       validate:"omitempty,dive"`
	StyleIDs       []int     `json:"style_ids" form:"style_ids"       validate:"omitempty,dive"`
}

type CatalogType string

var (
	CatalogCreated = CatalogType("created")
	CatalogUpdated = CatalogType("updated")
	CatalogDeleted = CatalogType("deleted")
)

type CatalogEvent struct {
	Type    CatalogType `json:"type"` // "created" | "updated" | "deleted"
	Payload uuid.UUID   `json:"payload"`
}

type CatalogItem struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Season   string   `json:"season,omitempty"`
	Styles   []string `json:"styles,omitempty"`
	Style    string   `json:"style,omitempty"`
	Gender   string   `json:"gender,omitempty"`
	ImageURL string   `json:"image_url,omitempty"`
}

type CatalogResponse struct {
	Items []CatalogItem `json:"items"`
	Total int           `json:"total"`
}
