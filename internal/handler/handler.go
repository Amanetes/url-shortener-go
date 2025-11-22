package handler

import (
	"url-shortener/internal/config"
	"url-shortener/internal/service"

	"github.com/go-playground/validator/v10"
)

type Handler struct {
	*service.Service
	cfg *config.Config
}

func New(services *service.Service, cfg *config.Config) *Handler {
	return &Handler{services, cfg}
}

type CustomValidator struct {
	v *validator.Validate
}

func NewValidator() *CustomValidator {
	return &CustomValidator{v: validator.New()}
}

func (cv *CustomValidator) Validate(i any) error {
	if err := cv.v.Struct(i); err != nil {
		return err
	}
	return nil
}
