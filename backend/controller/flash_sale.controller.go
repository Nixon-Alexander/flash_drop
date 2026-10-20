package controller

import (
	"database/sql"
	"errors"
	"strconv"

	"flash_drop/helper"
	"flash_drop/repository"
	"flash_drop/types"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FlashSaleController struct {
	logger *zap.Logger
	repo   *repository.FlashSaleRepository
}

func InitFlashSaleController(logger *zap.Logger, repo *repository.FlashSaleRepository) *FlashSaleController {
	return &FlashSaleController{logger: logger, repo: repo}
}

func (fc *FlashSaleController) CreateEvent(c *gin.Context) {
	var userInput types.UserInputFlashSalesType
	if err := c.ShouldBindJSON(&userInput); err != nil {
		fc.logger.Warn("invalid create event payload", zap.Error(err))
		helper.ErrorResponse(c, 400, err, "Invalid request body")
		return
	}

	if err := fc.repo.CreateEvent(c, userInput); err != nil {
		fc.logger.Error("failed to create event", zap.Error(err))
		helper.ErrorResponse(c, 500, err, "Error when creating event")
		return
	}

	helper.SuccessResponse(c, 201, nil, "Event created")
}

func (fc *FlashSaleController) GetAllEvent(c *gin.Context) {
	events, err := fc.repo.GetAllFlashSale(c)
	if err != nil {
		fc.logger.Error("failed to get events", zap.Error(err))
		helper.ErrorResponse(c, 500, err, "Error when getting events")
		return
	}

	helper.SuccessResponse(c, 200, events, "Events retrieved")
}

func (fc *FlashSaleController) GetEventById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helper.ErrorResponse(c, 400, err, "Invalid event id")
		return
	}

	event, err := fc.repo.GetFlashSaleById(c, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.ErrorResponse(c, 404, err, "Event not found")
			return
		}
		fc.logger.Error("failed to get event", zap.Int("id", id), zap.Error(err))
		helper.ErrorResponse(c, 500, err, "Error when getting event")
		return
	}

	helper.SuccessResponse(c, 200, event, "Event retrieved")
}

func (fc *FlashSaleController) DeleteEvent(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helper.ErrorResponse(c, 400, err, "Invalid event id")
		return
	}

	if err := fc.repo.DeleteEvent(c, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.ErrorResponse(c, 404, err, "Event not found")
			return
		}
		fc.logger.Error("failed to delete event", zap.Int("id", id), zap.Error(err))
		helper.ErrorResponse(c, 500, err, "Error when deleting event")
		return
	}

	helper.SuccessResponse(c, 200, nil, "Event deleted")
}

func (fc *FlashSaleController) UpdateStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		helper.ErrorResponse(c, 400, err, "Invalid event id")
		return
	}

	var body struct {
		Status types.FlashSaleStatus `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		helper.ErrorResponse(c, 400, err, "Invalid request body")
		return
	}

	switch body.Status {
	case types.SCHEDULED, types.ACTIVE, types.ENDED, types.CANCELLED:
	default:
		helper.ErrorResponse(c, 400, errors.New("invalid status"), "Status must be scheduled, active, ended, or cancelled")
		return
	}

	if err := fc.repo.UpdateStatus(c, id, body.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			helper.ErrorResponse(c, 404, err, "Event not found")
			return
		}
		fc.logger.Error("failed to update status", zap.Int("id", id), zap.Error(err))
		helper.ErrorResponse(c, 500, err, "Error when updating status")
		return
	}

	helper.SuccessResponse(c, 200, nil, "Status updated")
}
