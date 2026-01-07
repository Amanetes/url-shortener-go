package handler

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type HomeResponse struct {
	Message string `json:"message"`
	Version string `json:"version"`
	Time    string `json:"time"`
}

type HomeHandler struct{}

func NewHomeHandler() *HomeHandler {
	return &HomeHandler{}
}

func (h *HomeHandler) Home(c echo.Context) error {
	r := HomeResponse{
		Message: "URL Shortener API",
		Version: "1.0.0",
		Time:    time.Now().UTC().Format(time.RFC3339),
	}
	return c.JSON(http.StatusOK, r)
}
