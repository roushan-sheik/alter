package dto

import (
	"time"

	"gotickets/internal/querybuilder"

	"github.com/google/uuid"
)

type LibraryResponse struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	Category         string    `json:"category"`
	ShortDescription string    `json:"short_description"`
	ContentText      string    `json:"content_text"`
	ThumbnailURL     string    `json:"thumbnail_url"`
	MediaURL         *string   `json:"media_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PaginatedLibraryResponse struct {
	Data []*LibraryResponse `json:"data"`
	Meta *querybuilder.Meta `json:"meta"`
}

type LibraryDetailsResponse struct {
	LibraryItem *LibraryResponse   `json:"library_item"`
	Related     []*LibraryResponse `json:"related"`
}

type CategoryResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
