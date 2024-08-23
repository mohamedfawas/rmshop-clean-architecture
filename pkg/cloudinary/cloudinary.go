package cloudinary

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryService struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinaryService(cloudName, apiKey, apiSecret string) (*CloudinaryService, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
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
		return "", err
	}

	return result.SecureURL, nil
}

func (s *CloudinaryService) DeleteImage(ctx context.Context, publicID string) error {
	_, err := s.cld.Upload.Destroy(ctx, uploader.DestroyParams{PublicID: publicID})
	return err
}

func generateUniqueID() string {
	// Implement a function to generate a unique ID (you can use UUID or any other method)
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
