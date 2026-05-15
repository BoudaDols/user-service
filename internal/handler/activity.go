package handler

import (
	"net/http"
	"strconv"

	"github.com/dolsom/user-service/internal/service"
	"github.com/gin-gonic/gin"
)

type ActivityHandler struct {
	svc *service.ActivityService
}

func NewActivityHandler(svc *service.ActivityService) *ActivityHandler {
	return &ActivityHandler{svc: svc}
}

func (h *ActivityHandler) GetByUserID(c *gin.Context) {
	userID := c.Param("user_id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.svc.GetByUserID(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "internal error"})
		return
	}

	if logs == nil {
		logs = []any{}
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": logs})
}
