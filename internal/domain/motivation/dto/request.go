package dto

type CreateMotivationReq struct {
	Title        string `json:"title" validate:"required,max=255"`
	SpeakerName  string `json:"speaker_name" validate:"required,max=255"`
	Description  string `json:"description" validate:"required"`
	VideoURL     string `json:"video_url" validate:"required"`
	ThumbnailURL string `json:"thumbnail_url" validate:"required"`
	Duration     string `json:"duration" validate:"required,max=20"`
}

type UpdateMotivationReq struct {
	Title        string `json:"title" validate:"required,max=255"`
	SpeakerName  string `json:"speaker_name" validate:"required,max=255"`
	Description  string `json:"description" validate:"required"`
	VideoURL     string `json:"video_url" validate:"required"`
	ThumbnailURL string `json:"thumbnail_url" validate:"required"`
	Duration     string `json:"duration" validate:"required,max=20"`
}
