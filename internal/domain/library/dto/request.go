package dto

type CreateLibraryReq struct {
	Title            string  `json:"title" validate:"required,max=255"`
	Category         string  `json:"category" validate:"required,max=100"`
	ShortDescription string  `json:"short_description" validate:"required"`
	ContentText      string  `json:"content_text" validate:"required"`
	ThumbnailURL     string  `json:"thumbnail_url" validate:"required,url"`
	MediaURL         *string `json:"media_url" validate:"omitempty,url"`
}

type UpdateLibraryReq struct {
	Title            string  `json:"title" validate:"required,max=255"`
	Category         string  `json:"category" validate:"required,max=100"`
	ShortDescription string  `json:"short_description" validate:"required"`
	ContentText      string  `json:"content_text" validate:"required"`
	ThumbnailURL     string  `json:"thumbnail_url" validate:"required,url"`
	MediaURL         *string `json:"media_url" validate:"omitempty,url"`
}

type CreateCategoryReq struct {
	Name string `json:"name" validate:"required,max=100"`
}

type UpdateCategoryReq struct {
	Name string `json:"name" validate:"required,max=100"`
}
