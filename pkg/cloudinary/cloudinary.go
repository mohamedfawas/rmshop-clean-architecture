package cloudinary

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"time"
	"net/url"
	"strings"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudName, apiKey, apiSecret string) (*CloudinaryService, error) {
	log.Printf("Cloudinary config - cloudName=%s", cloudName)
	log.Printf("Cloudinary config - apiKey=%s", apiKey)
	log.Printf("Cloudinary config - apiSecret=%s", apiSecret)
	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary credentials are incomplete")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		log.Printf("Failed to create cloudinary service instance : %v", err)
		return nil, err
	}
	return &CloudinaryService{cld: cld}, nil
}

func (s *CloudinaryService) UploadImage(ctx context.Context, file multipart.File, productSlug string) (string, error) {
	uploadParams := uploader.UploadParams{
		PublicID: productSlug + "_" + generateUniqueID(),
		Folder:   "product_images",
	}

	result, err := s.cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		log.Printf("error while uploading to cloudinary :%v", err)
		return "", err
	}
	log.Printf("Cloudinary raw result: %#v", result)
	if result.Error.Message != "" {
		log.Printf("Cloudinary logical error: %s", result.Error.Message)
		return "", fmt.Errorf("cloudinary error: %s", result.Error.Message)
	}

	if result.SecureURL == "" {
		log.Printf("Cloudinary returned empty secure_url")
		return "", fmt.Errorf("cloudinary upload returned empty secure_url")
	}

	log.Printf("Cloudinary upload success: public_id=%s, url=%s", result.PublicID, result.SecureURL)

	return result.SecureURL, nil
}

func (s *CloudinaryService) DeleteImage(ctx context.Context, imageURL string) error {
    publicID := extractPublicID(imageURL)

    res, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
    if err != nil {
        log.Printf("Cloudinary destroy error (public_id=%s): %v", publicID, err)
        return err
    }
    log.Printf("Cloudinary destroy result: public_id=%s, result=%s", publicID, res.Result)
    return nil
}

func generateUniqueID() string {
	// Implement a function to generate a unique ID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}



// extractPublicID turns a Cloudinary URL into the public_id expected by Destroy.
func extractPublicID(imageURL string) string {
    u, err := url.Parse(imageURL)
    if err != nil {
        return imageURL // fallback
    }

    parts := strings.Split(u.Path, "/")
    // path looks like: /<cloud>/image/upload/v123456789/product_images/....png
    // find "upload"
    for i, p := range parts {
        if p == "upload" && i+1 < len(parts) {
            // after "upload" we may have "v123..." then the actual path
            rest := parts[i+1+1:] // skip "upload" and version "v..."
            if len(rest) == 0 {
                return imageURL
            }
            idWithExt := strings.Join(rest, "/")
            if dot := strings.LastIndex(idWithExt, "."); dot > 0 {
                return idWithExt[:dot] // strip extension
            }
            return idWithExt
        }
    }
    return imageURL
}