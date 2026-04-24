package service

import (
	"context"
	"fmt"
	"time"
	"net/http"
	"enfok_bd/src/utils"

	"enfok_bd/src/domain/image"
	"enfok_bd/src/domain/ports/driven"
	"enfok_bd/src/domain/ports/driving"

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

func (s *ImageService) UpdateResult(ctx context.Context, uuid string, content []byte, nodeID string) (string, error) {
	mimeType := http.DetectContentType(content)
	extension := utils.GetExtensionFromMimeType(mimeType)

	fileName := fmt.Sprintf("result_%s%s", uuid, extension)

	storagePath, err := s.storageRepo.Upload(ctx, fileName, content)
	if err != nil {
		return "", fmt.Errorf("failed to upload result to storage: %w", err)
	}

	// 4. Actualizar la base de datos con la nueva URL y el ID del nodo
	if err := s.imageRepo.UpdateResult(ctx, uuid, storagePath, nodeID); err != nil {
		return "", fmt.Errorf("failed to update result in database: %w", err)
	}

	return storagePath, nil
}

func (s *ImageService) GetImage(ctx context.Context, uuid string) (*image.Image, error) {
	return s.imageRepo.GetByID(ctx, uuid)
}

// UploadBatch crea un batch y sube múltiples imágenes bajo el mismo batch_uuid.
// Retorna: batchUUID, []imageUUIDs, error
func (s *ImageService) UploadBatch(ctx context.Context, userUUID string, files []driving.BatchFile) (string, []string, error) {
	batchUUID := uuid.New().String()
	batch := &image.Batch{
		BatchUUID:   batchUUID,
		UserUUID:    userUUID,
		RequestTime: time.Now(),
		Status:      "PENDING",
	}
	if err := s.batchRepo.Save(ctx, batch); err != nil {
		return "", nil, fmt.Errorf("failed to create batch: %w", err)
	}

	imageUUIDs := make([]string, 0, len(files))
	for _, f := range files {
		imageUUID := uuid.New().String()
		storagePath, err := s.storageRepo.Upload(ctx, imageUUID+"_"+f.FileName, f.Content)
		if err != nil {
			return "", nil, fmt.Errorf("failed to upload file %s: %w", f.FileName, err)
		}

		img := &image.Image{
			ImageUUID:     imageUUID,
			BatchUUID:     batchUUID,
			OriginalName:  f.FileName,
			ResultPath:    &storagePath,
			Status:        "RECEIVED",
			ReceptionTime: time.Now(),
		}
		if err := s.imageRepo.Save(ctx, img); err != nil {
			return "", nil, fmt.Errorf("failed to save image record %s: %w", f.FileName, err)
		}
		imageUUIDs = append(imageUUIDs, imageUUID)
	}

	return batchUUID, imageUUIDs, nil
}

func (s *ImageService) RegisterImage(ctx context.Context, userUUID, batchUUID, imageUUID, fileName string) error {
	img := &image.Image{
		ImageUUID:     imageUUID,
		BatchUUID:     batchUUID,
		OriginalName:  fileName,
		Status:        "RECEIVED",
		ReceptionTime: time.Now(),
	}
	return s.imageRepo.Save(ctx, img)
}

func (s *ImageService) RegisterBatch(ctx context.Context, userUUID, batchUUID string) error {
	batch := &image.Batch{
		BatchUUID:   batchUUID,
		UserUUID:    userUUID,
		RequestTime: time.Now(),
		Status:      "PENDING",
	}
	return s.batchRepo.Save(ctx, batch)
}

func (s *ImageService) GetUserStatistics(ctx context.Context, userUUID string) (*driving.UserStatistics, error) {
	stats, err := s.imageRepo.GetUserStatistics(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	topTrans := make([]driving.TransformationStat, 0, len(stats.TopTransformations))
	for _, t := range stats.TopTransformations {
		topTrans = append(topTrans, driving.TransformationStat{
			Name:  t.Name,
			Count: t.Count,
		})
	}

	return &driving.UserStatistics{
		TotalBatches:       stats.TotalBatches,
		TotalImages:        stats.TotalImages,
		ImagesSuccess:      stats.ImagesSuccess,
		ImagesFailed:       stats.ImagesFailed,
		TopTransformations: topTrans,
	}, nil
}

func (s *ImageService) GetUserActivity(ctx context.Context, userUUID string) ([]driving.UserActivity, error) {
	activities, err := s.imageRepo.GetUserActivity(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	result := make([]driving.UserActivity, 0, len(activities))
	for _, a := range activities {
		result = append(result, driving.UserActivity{
			EventType:   a.EventType,
			RefUUID:     a.RefUUID,
			ParentUUID:  a.ParentUUID,
			Description: a.Description,
			OccurredAt:  a.OccurredAt,
		})
	}
	return result, nil
}

func (s *ImageService) GetBatchImagesPaginated(ctx context.Context, batchUUID string, page, limit int) (*driving.PaginatedImages, error) {
	offset := (page - 1) * limit
	images, total, err := s.imageRepo.ListByBatchPaginated(ctx, batchUUID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Generar URLs para cada imagen
	for i := range images {
		if images[i].ResultPath != nil && *images[i].ResultPath != "" {
			url, err := s.storageRepo.GetPresignedURL(ctx, *images[i].ResultPath, time.Hour)
			if err == nil {
				images[i].ResultPath = &url
			}
		}
	}

	hasMore := (offset + limit) < total

	return &driving.PaginatedImages{
		Images:      images,
		CurrentPage: page,
		Limit:       limit,
		TotalCount:  total,
		HasMore:     hasMore,
	}, nil
}

func (s *ImageService) DownloadImage(ctx context.Context, uuid string) ([]byte, string, error) {
	img, err := s.imageRepo.GetByID(ctx, uuid)
	if err != nil {
		return nil, "", fmt.Errorf("image not found: %w", err)
	}

	if img.ResultPath == nil || *img.ResultPath == "" {
		return nil, "", fmt.Errorf("image has no result path yet")
	}

	content, err := s.storageRepo.Download(ctx, *img.ResultPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download from storage: %w", err)
	}

	return content, img.OriginalName, nil
}
