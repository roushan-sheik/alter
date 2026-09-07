package dto

import (
	"time"

	"gotickets/internal/querybuilder"

	"github.com/google/uuid"
)

type MotivationResponse struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	SpeakerName  string    `json:"speaker_name"`
	Description  string    `json:"description"`
	VideoURL     string    `json:"video_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	Duration     string    `json:"duration"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PaginatedMotivationResponse is returned by the list endpoint.
type PaginatedMotivationResponse struct {
	Data []*MotivationResponse `json:"data"`
	Meta *querybuilder.Meta    `json:"meta"`
}

// MotivationDetailsResponse is returned by the details endpoint.
type MotivationDetailsResponse struct {
	Motivation *MotivationResponse   `json:"motivation"`
	Related    []*MotivationResponse `json:"related"`
}
