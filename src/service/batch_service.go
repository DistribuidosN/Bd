package service

import (
	"context"
	"time"

	"enfok_bd/src/domain/image"
	"enfok_bd/src/domain/ports/driven"
	"enfok_bd/src/domain/ports/driving"
)

type BatchService struct {
	repo        driven.BatchRepository
	imgRepo     driven.ImageRepository
	storageRepo driven.StorageRepository
}

func NewBatchService(r driven.BatchRepository, ir driven.ImageRepository, storage driven.StorageRepository) *BatchService {
	return &BatchService{repo: r, imgRepo: ir, storageRepo: storage}
}

func (s *BatchService) GetBatch(ctx context.Context, uuid string) (*image.Batch, error) {
	return s.repo.GetByID(ctx, uuid)
}

func (s *BatchService) UpdateStatus(ctx context.Context, uuid string, status string) error {
	return s.repo.UpdateStatus(ctx, uuid, status)
}

// GetBatchWithImages devuelve el batch junto con todas sus imágenes asociadas.
func (s *BatchService) GetBatchWithImages(ctx context.Context, uuid string) (*image.Batch, []image.Image, error) {
	batch, err := s.repo.GetByID(ctx, uuid)
	if err != nil {
		return nil, nil, err
	}
	images, err := s.repo.GetImagesByBatch(ctx, uuid)
	if err != nil {
		return nil, nil, err
	}
	return batch, images, nil
}

func (s *BatchService) SaveTransformations(ctx context.Context, batchUUID string, transformations []image.BatchTransformation) error {
	return s.repo.SaveTransformations(ctx, batchUUID, transformations)
}

func (s *BatchService) ListUserBatches(ctx context.Context, userUUID string) ([]driving.BatchWithCover, error) {
	batches, err := s.repo.ListByUserWithCover(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	result := make([]driving.BatchWithCover, 0, len(batches))
	for _, b := range batches {
		coverURL := ""
		if b.CoverResultPath != nil && *b.CoverResultPath != "" {
			url, err := s.storageRepo.GetPresignedURL(ctx, *b.CoverResultPath, time.Hour)
			if err == nil {
				coverURL = url
			}
		}

		coverUUID := ""
		if b.CoverImageUUID != nil {
			coverUUID = *b.CoverImageUUID
		}

		result = append(result, driving.BatchWithCover{
			Batch:          b.Batch,
			CoverImageURL:  coverURL,
			CoverImageUUID: coverUUID,
		})
	}
	return result, nil
}
func (s *BatchService) GetProgress(ctx context.Context, batchUUID string) (*driving.PaginatedImagesProgress, error) {
	p, err := s.imgRepo.GetBatchProgress(ctx, batchUUID)
	if err != nil {
		return nil, err
	}
	return &driving.PaginatedImagesProgress{
		BatchUUID:          p.BatchUUID,
		TotalImages:        p.TotalImages,
		ProcessedImages:    p.ProcessedImages,
		ProgressPercentage: p.ProgressPercentage,
	}, nil
}
