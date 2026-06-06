package service

import "volunteer-platform/internal/repository"

type SaveCroppedImage func(dataURL string) (string, error)
type SaveUploadedImage func() (string, error)

type EventLogic interface {
	ResolveCover(currentCover, croppedData string, saveCropped SaveCroppedImage, saveUpload SaveUploadedImage) string
}

type Events struct{}

// ResolveCover выбирает новую обложку: cropped image, upload или текущая.
func (Events) ResolveCover(currentCover, croppedData string, saveCropped SaveCroppedImage, saveUpload SaveUploadedImage) string {
	if croppedData != "" {
		if path, err := saveCropped(croppedData); err == nil && path != "" {
			return path
		}
	}
	if saveUpload != nil {
		if path, err := saveUpload(); err == nil && path != "" {
			return path
		}
	}
	return currentCover
}

type Services struct {
	Store  repository.CRUD
	Events EventLogic
}

// New собирает HTTP router поверх готового набора handlers.
func New(store repository.CRUD) *Services { return &Services{Store: store, Events: Events{}} }

// NewServices создаёт сервисы без repository для legacy tests.
func NewServices() *Services { return New(nil) }
