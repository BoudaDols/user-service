package handler

import (
	"errors"
	"net/http"

	"github.com/dolsom/user-service/internal/service"
	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	svc *service.ProfileService
}

func NewProfileHandler(svc *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{svc: svc}
}

func (h *ProfileHandler) Get(c *gin.Context) {
	userID := c.Param("user_id")

	profile, err := h.svc.Get(userID)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}

func (h *ProfileHandler) Create(c *gin.Context) {
	var body struct {
		UserID string `json:"user_id" binding:"required"`
		Email  string `json:"email"   binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	profile, err := h.svc.Create(body.UserID, body.Email)
	if err != nil {
		if errors.Is(err, service.ErrProfileAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"success": false, "message": "profile already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": profile})
}

func (h *ProfileHandler) Update(c *gin.Context) {
	userID := c.Param("user_id")

	var body map[string]any
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	// Only allow safe fields to be updated
	allowed := map[string]bool{"display_name": true, "avatar_url": true, "bio": true, "language": true, "timezone": true}
	fields := map[string]any{}
	for k, v := range body {
		if allowed[k] {
			fields[k] = v
		}
	}

	profile, err := h.svc.Update(userID, fields)
	if err != nil {
		if errors.Is(err, service.ErrProfileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "message": "profile not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}
