package http

import (
	"net/http"

	"datacenter-calc/internal/repo"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	CalcRepo *repo.CalculationRepository
}

func NewCartHandler(calcRepo *repo.CalculationRepository) *CartHandler {
	return &CartHandler{CalcRepo: calcRepo}
}

func (h *CartHandler) GetCartInfo(c *gin.Context) {
	draft := h.CalcRepo.GetDraftByUser(CurrentUserID)
	if draft == nil {
		draft = h.CalcRepo.CreateDraft(CurrentUserID)
	}

	count := h.CalcRepo.GetTotalItemsInDraft(CurrentUserID)

	c.JSON(http.StatusOK, gin.H{
		"calculation_id": draft.ID,
		"total_items":    count,
	})
}
