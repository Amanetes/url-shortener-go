package handler

import (
	"fmt"
	"net/http"
	"url-shortener/internal/config"
	"url-shortener/internal/service"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type (
	UrlRequest struct {
		LongUrl string `json:"long_url" validate:"required"`
	}

	UrlResponse struct {
		Code     string `json:"code"`
		ShortUrl string `json:"short_url"`
	}

	URLHandler struct {
		cfg     *config.Config
		service service.URLServicer
		v       *validator.Validate
	}
)

func NewURLHandler(service service.URLServicer, cfg *config.Config, v *validator.Validate) *URLHandler {
	return &URLHandler{
		service: service,
		cfg:     cfg,
		v:       v,
	}
}

func (h *URLHandler) Shorten(c echo.Context) error {
	var req UrlRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	if err := h.v.Struct(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	u, err := h.service.Create(c.Request().Context(), req.LongUrl)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to shorten url")
	}

	scheme := c.Scheme()

	// Собираем полный URL
	fullShortUrl := fmt.Sprintf("%s://%s/urls/%s", scheme, h.cfg.App.Name, u.Code)

	resp := UrlResponse{
		Code:     u.Code,
		ShortUrl: fullShortUrl,
	}

	return c.JSON(http.StatusCreated, resp)
}

func (h *URLHandler) Redirect(c echo.Context) error {
	code := c.Param("code")
	if code == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "code is required")
	}

	u, err := h.service.GetByCode(c.Request().Context(), code)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "url not found")
	}

	return c.Redirect(http.StatusTemporaryRedirect, u.LongUrl)
}
