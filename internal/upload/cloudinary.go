package upload

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"path"
	"regexp"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// versionSegment matches a Cloudinary version prefix like "v1787740828/".
var versionSegment = regexp.MustCompile(`^v\d+/`)

// PublicIDFromURL extracts the Cloudinary public_id from a full secure URL.
//
// Example:
//
//	"https://res.cloudinary.com/demo/image/upload/v1234567890/zick/avatars/abc.jpg"
//	→ "zick/avatars/abc"
func PublicIDFromURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Path looks like: /demo/image/upload/v1234567890/zick/avatars/abc.jpg
	parts := strings.SplitN(u.Path, "/upload/", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("URL does not look like a Cloudinary upload URL")
	}

	after := parts[1]                      // "v1234567890/zick/avatars/abc.jpg"
	after = versionSegment.ReplaceAllString(after, "") // "zick/avatars/abc.jpg"
	after = strings.TrimSuffix(after, path.Ext(after)) // "zick/avatars/abc"
	return after, nil
}


type cloudinaryUploader struct {
	cld *cloudinary.Cloudinary
}

// NewCloudinaryUploader initializes the Cloudinary SDK with the given credentials.
func NewCloudinaryUploader(cloudName, apiKey, apiSecret string) (Uploader, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %w", err)
	}
	return &cloudinaryUploader{cld: cld}, nil
}

func (cu *cloudinaryUploader) Upload(ctx context.Context, file multipart.File, folder string) (UploadResult, error) {
	resp, err := cu.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return UploadResult{}, fmt.Errorf("cloudinary upload error: %w", err)
	}

	return UploadResult{
		URL:      resp.SecureURL,
		PublicID: resp.PublicID,
	}, nil
}

func (cu *cloudinaryUploader) Delete(ctx context.Context, publicID string) error {
	_, err := cu.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("cloudinary delete error: %w", err)
	}
	return nil
}
