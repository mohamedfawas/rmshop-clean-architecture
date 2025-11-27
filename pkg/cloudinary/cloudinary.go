package cloudinary

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudName, apiKey, apiSecret string) (*CloudinaryService, error) {
	log.Printf("Cloudinary config - cloudName=%s", cloudName)
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

func (s *CloudinaryService) DeleteImage(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}

func generateUniqueID() string {
	// Implement a function to generate a unique ID
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
