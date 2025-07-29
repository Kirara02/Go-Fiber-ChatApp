package services

import (
	"context"
	"main/config"
	"mime/multipart"
	"path"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type UploadService interface {
	UploadFile(file *multipart.FileHeader, subfolder string, publicID string) (string, error)
	DeleteFile(publicID string) error
}

type uploadService struct {
	cfg        *config.Config
	baseFolder string
}

func NewUploadService(cfg *config.Config) UploadService {
	return &uploadService{
		cfg:        cfg,
		baseFolder: cfg.CloudinaryBaseFolder,
	}
}

func (s *uploadService) UploadFile(file *multipart.FileHeader, subfolder string, publicID string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	cld, err := cloudinary.NewFromParams(
		s.cfg.CloudinaryCloudName,
		s.cfg.CloudinaryAPIKey,
		s.cfg.CloudinaryAPISecret,
	)
	if err != nil {
		return "", err
	}

	overwrite := true
	uploadParams := uploader.UploadParams{
		PublicID:  publicID,
		Folder:    path.Join(s.baseFolder, subfolder),
		Overwrite: &overwrite,
	}

	uploadResult, err := cld.Upload.Upload(context.Background(), src, uploadParams)
	if err != nil {
		return "", err
	}
	return uploadResult.SecureURL, nil
}

func (s *uploadService) DeleteFile(publicID string) error {
	cld, err := cloudinary.NewFromParams(
		s.cfg.CloudinaryCloudName,
		s.cfg.CloudinaryAPIKey,
		s.cfg.CloudinaryAPISecret,
	)
	if err != nil {
		return err
	}

	_, err = cld.Upload.Destroy(context.Background(), uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}
