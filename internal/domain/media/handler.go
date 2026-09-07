package media

import (
	"net/http"

	"gotickets/internal/httpresponse"
	"gotickets/internal/upload"

	"github.com/labstack/echo/v5"
)


// UploadResponse is returned by POST /api/v1/upload.
type UploadResponse struct {
	URL      string `json:"url"       example:"https://res.cloudinary.com/demo/image/upload/sample.jpg"`
	PublicID string `json:"public_id" example:"zick/content/abc123"`
}

// DeleteRequest is the body for DELETE /api/v1/upload.
// Pass the full Cloudinary URL — the server extracts the public_id automatically.
type DeleteRequest struct {
	URL string `json:"url" validate:"required,url" example:"https://res.cloudinary.com/demo/image/upload/v1234/zick/uploads/abc.jpg"`
}

// DeleteResponse is returned by DELETE /api/v1/upload.
type DeleteResponse struct {
	Message string `json:"message" example:"File deleted successfully"`
}

// Handler holds the Echo HTTP handlers for the media domain.
type Handler struct {
	uploader        upload.Uploader
	maxUploadSizeMB int
}

// NewHandler creates a new media Handler.
func NewHandler(uploader upload.Uploader, maxUploadSizeMB int) *Handler {
	if maxUploadSizeMB <= 0 {
		maxUploadSizeMB = 200 // default to 200MB if not set
	}
	return &Handler{uploader: uploader, maxUploadSizeMB: maxUploadSizeMB}
}

// Upload godoc
// @Summary      Upload a file
// @Description  Uploads any file (image, audio, etc.) to Cloudinary and returns the public URL.
// @Description  Pass an optional `folder` query parameter to organise assets (e.g. `?folder=content/thumbnails`).
// @Description  The returned `url` can be stored and referenced by any domain (content, profile, etc.).
// @Tags         Media
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file    formData  file    true   "File to upload"
// @Param        folder  query     string  false  "Cloudinary folder (default: zick/uploads)"
// @Success      200     {object}  UploadResponse
// @Failure      400     {object}  httpresponse.Error  "Missing or oversized file"
// @Failure      401     {object}  httpresponse.Error
// @Failure      503     {object}  httpresponse.Error  "Upload service unavailable"
// @Router       /api/v1/upload [post]
func (h *Handler) Upload(c *echo.Context) error {
	if h.uploader == nil {
		return c.JSON(http.StatusServiceUnavailable, httpresponse.NewError(
			http.StatusServiceUnavailable, "Upload service not configured", "",
		))
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(
			http.StatusBadRequest, "File is required", err.Error(),
		))
	}

	maxBytes := int64(h.maxUploadSizeMB) << 20
	if fileHeader.Size > maxBytes {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(
			http.StatusBadRequest, "File too large", "",
		))
	}

	folder := c.QueryParam("folder")
	if folder == "" {
		folder = "zick/uploads"
	}

	file, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(
			http.StatusInternalServerError, "Failed to read file", err.Error(),
		))
	}
	defer file.Close()

	result, err := h.uploader.Upload(c.Request().Context(), file, folder)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(
			http.StatusInternalServerError, "Upload failed", err.Error(),
		))
	}

	return c.JSON(http.StatusOK, UploadResponse{
		URL:      result.URL,
		PublicID: result.PublicID,
	})
}

// Delete godoc
// @Summary      Delete an uploaded file
// @Description  Permanently removes a file from Cloudinary using its public_id.
// @Description  The public_id is returned by POST /api/v1/upload.
// @Tags         Media
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      DeleteRequest        true  "Full Cloudinary URL to delete"
// @Success      200      {object}  DeleteResponse
// @Failure      400      {object}  httpresponse.Error   "Missing public_id"
// @Failure      401      {object}  httpresponse.Error
// @Failure      503      {object}  httpresponse.Error   "Upload service unavailable"
// @Router       /api/v1/upload [delete]
func (h *Handler) Delete(c *echo.Context) error {
	if h.uploader == nil {
		return c.JSON(http.StatusServiceUnavailable, httpresponse.NewError(
			http.StatusServiceUnavailable, "Upload service not configured", "",
		))
	}

	var req DeleteRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(
			http.StatusBadRequest, "Invalid request body", err.Error(),
		))
	}
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(
			http.StatusBadRequest, "url is required", err.Error(),
		))
	}

	publicID, err := upload.PublicIDFromURL(req.URL)
	if err != nil {
		return c.JSON(http.StatusBadRequest, httpresponse.NewError(
			http.StatusBadRequest, "Invalid Cloudinary URL", err.Error(),
		))
	}

	if err := h.uploader.Delete(c.Request().Context(), publicID); err != nil {
		return c.JSON(http.StatusInternalServerError, httpresponse.NewError(
			http.StatusInternalServerError, "Failed to delete file", err.Error(),
		))
	}

	return c.JSON(http.StatusOK, DeleteResponse{Message: "File deleted successfully"})
}
