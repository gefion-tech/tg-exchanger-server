package ma

import (
	"net/http"
	"reflect"

	AppError "github.com/gefion-tech/tg-exchanger-server/internal/core/errors"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gin-gonic/gin"
)

// Универсальный метод выполнения CRUD операций
func (m *ModMerchantAutoPayout) MerchantAutopayoutHandler(c *gin.Context) {
	var r models.MerchantAutopayout
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&r); err != nil {
			m.responser.Error(c, http.StatusUnprocessableEntity, AppError.ErrInvalidBody)
			return
		}

		if err := r.Validation(); err != nil {
			m.responser.Error(c, http.StatusUnprocessableEntity, err)
			return
		}
	}

	if obj := m.responser.RecordHandler(c, &r); obj != nil {
		if reflect.TypeOf(obj) != reflect.TypeOf(&models.MerchantAutopayout{}) {
			return
		}

		switch c.Request.Method {
		case http.MethodPost:
			m.responser.CreateRecordResponse(c, m.repository.MerchantAutopayout(), obj)
			return
		case http.MethodGet:
			m.responser.GetRecordResponse(c, m.repository.MerchantAutopayout(), obj)
			return
		case http.MethodPut:
			m.responser.UpdateRecordResponse(c, m.repository.MerchantAutopayout(), obj)
			return
		case http.MethodDelete:
			m.responser.DeleteRecordResponse(c, m.repository.MerchantAutopayout(), obj)
			return
		}
	}

	m.responser.Error(c, http.StatusInternalServerError, AppError.ErrFailedToInitializeStruct)
}
