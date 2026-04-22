package service

import (
	"context"
	"time"

	"enfok_bd/src/domain/image"
	"enfok_bd/src/domain/ports/driven"

	"github.com/google/uuid"
)

type ImageService struct {
	imageRepo   driven.ImageRepository
	storageRepo driven.StorageRepository
	batchRepo   driven.BatchRepository
}

func NewImageService(ir driven.ImageRepository, sr driven.StorageRepository, br driven.BatchRepository) *ImageService {
	return &ImageService{
		imageRepo:   ir,
		storageRepo: sr,
		batchRepo:   br,
	}
}

func (s *ImageService) ProcessUpload(ctx context.Context, userUUID, fileName string, content []byte) (string, error) {
	batchUUID := uuid.New().String()
	batch := &image.Batch{
		BatchUUID:   batchUUID,
		UserUUID:    userUUID,
		RequestTime: time.Now(),
		Status:      "PENDING",
	}
	if err := s.batchRepo.Save(ctx, batch); err != nil {
		return "", err
	}

	imageUUID := uuid.New().String()
	storagePath, err := s.storageRepo.Upload(ctx, imageUUID+"_"+fileName, content)
	if err != nil {
		return "", err
	}

	img := &image.Image{
		ImageUUID:     imageUUID,
		BatchUUID:     batchUUID,
		OriginalName:  fileName,
		ResultPath:    &storagePath,
		Status:        "RECEIVED",
		ReceptionTime: time.Now(),
	}
	if err := s.imageRepo.Save(ctx, img); err != nil {
		return "", err
	}

	return imageUUID, nil
}

func (s *ImageService) UpdateStatus(ctx context.Context, uuid string, status string) error {
	return s.imageRepo.UpdateStatus(ctx, uuid, status)
}

func (s *ImageService) GetImage(ctx context.Context, uuid string) (*image.Image, error) {
	return s.imageRepo.GetByID(ctx, uuid)
}
