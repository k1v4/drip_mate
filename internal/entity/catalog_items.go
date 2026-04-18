package entity

import (
	"time"

	"github.com/google/uuid"
)

type Catalog struct {
	ID             uuid.UUID `db:"id" json:"id"`
	Name           string    `db:"name" json:"name"`
	CategoryID     int       `db:"category_id" json:"category_id"`
	Gender         *string   `db:"gender" json:"gender"`
	SeasonID       int       `db:"season_id" json:"season_id"`
	FormalityLevel *int16    `db:"formality_level" json:"formality_level"`
	Material       *string   `db:"material" json:"material"`
	ImageURL       string    `db:"image_url" json:"image_url"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
	IsDeleted      bool      `db:"is_deleted" json:"is_deleted"`
}
