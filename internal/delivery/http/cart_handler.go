package http

import (
	"net/http"

	"datacenter-calc/internal/auth"
	"datacenter-calc/internal/repo"

	"github.com/gin-gonic/gin"
)

type CartHandler struct {
	CalcRepo *repo.CalculationRepository
}

func NewCartHandler(calcRepo *repo.CalculationRepository) *CartHandler {
	return &CartHandler{CalcRepo: calcRepo}
}

// GetCartInfo godoc
// @Summary Получить информацию о корзине
// @Description Возвращает ID черновика и количество услуг в нём
// @Tags cart
// @Security Bearer
// @Produce json
// @Success 200 {object} CartResponse
// @Failure 401 {object} ErrorResponse
// @Router /cart [get]
func (h *CartHandler) GetCartInfo(c *gin.Context) {
	userID := auth.UserIDFromContext(c)

	draft := h.CalcRepo.GetDraftByUser(userID)
	if draft == nil {
		draft = h.CalcRepo.CreateDraft(userID)
	}

	count := h.CalcRepo.GetTotalItemsInDraft(userID)

	c.JSON(http.StatusOK, gin.H{
		"calculation_id": draft.ID,
		"total_items":    count,
	})
}
