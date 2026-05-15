package handler

import (
	"net/http"

	"github.com/dolsom/user-service/internal/model"
	"github.com/dolsom/user-service/internal/service"
	"github.com/gin-gonic/gin"
)

type PreferenceHandler struct {
	svc *service.PreferenceService
}

func NewPreferenceHandler(svc *service.PreferenceService) *PreferenceHandler {
	return &PreferenceHandler{svc: svc}
}

func (h *PreferenceHandler) GetAll(c *gin.Context) {
	userID := c.Param("user_id")

	prefs, err := h.svc.GetAll(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
		return
	}

	if prefs == nil {
		prefs = []model.Preference{}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": prefs})
}

func (h *PreferenceHandler) Upsert(c *gin.Context) {
	userID := c.Param("user_id")

	var body struct {
		Key   string `json:"key"   binding:"required"`
		Value string `json:"value" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "message": err.Error()})
		return
	}

	if err := h.svc.Upsert(userID, body.Key, body.Value); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "preference saved"})
}

func (h *PreferenceHandler) Delete(c *gin.Context) {
	userID := c.Param("user_id")
	key := c.Param("key")

	if err := h.svc.Delete(userID, key); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "preference deleted"})
}
