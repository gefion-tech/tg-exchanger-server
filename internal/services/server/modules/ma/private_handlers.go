package ma

import (
	"encoding/json"
	"net/http"
	"reflect"

	AppError "github.com/gefion-tech/tg-exchanger-server/internal/core/errors"
	AppType "github.com/gefion-tech/tg-exchanger-server/internal/core/types"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gin-gonic/gin"
)

/*
	@Method POST
	@Path admin/merchant-autopayout
	@Type PRIVATE
	@Documentation

	Создать запись в таблице `merchant_autopayout`

	# TESTED
*/
func (m *ModMerchantAutoPayout) CreateMerchantAutopayoutHandler(c *gin.Context) {
	m.MerchantAutopayoutHandler(c)
}

/*
	@Method PUT
	@Path admin/merchant-autopayout/:id
	@Type PRIVATE
	@Documentation

	Обновить запись в таблице `merchant_autopayout`

	# TESTED
*/
func (m *ModMerchantAutoPayout) UpdateMerchantAutopayoutHandler(c *gin.Context) {
	m.MerchantAutopayoutHandler(c)
}

/*
	@Method DELETE
	@Path admin/merchant-autopayout/:id
	@Type PRIVATE
	@Documentation

	Удалить запись в таблице `merchant_autopayout`

	# TESTED
*/
func (m *ModMerchantAutoPayout) DeleteMerchantAutopayoutHandler(c *gin.Context) {
	m.MerchantAutopayoutHandler(c)
}

/*
	@Method GET
	@Path admin/merchant-autopayout/:id
	@Type PRIVATE
	@Documentation

	Получить запись из таблицы `merchant_autopayout`

	# TESTED
*/
func (m *ModMerchantAutoPayout) GetMerchantAutopayoutHandler(c *gin.Context) {
	m.MerchantAutopayoutHandler(c)
}

/*
	@Method GET
	@Path admin/merchant-autopayout/:id
	@Type PRIVATE
	@Documentation

	Получение лимитированного объема записей из таблицы `merchant_autopayout`

	# TESTED
*/
func (m *ModMerchantAutoPayout) GetMerchantAutopayoutSelectionHandler(c *gin.Context) {
	s := &models.MerchantAutopayoutSelection{
		Service: []string{c.Query("service")},
	}

	m.responser.SelectionResponse(c, m.repository, s)
}

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

func (m *ModMerchantAutoPayout) PingMerchantAutopayoutHandler(c *gin.Context) {
	var r models.MerchantAutopayout
	if obj := m.responser.RecordHandler(c, &r); obj != nil {
		if reflect.TypeOf(obj) != reflect.TypeOf(&models.MerchantAutopayout{}) {
			return
		}

		// Достаю из БД нужную запись
		if m.repository.MerchantAutopayout().Get(obj.(*models.MerchantAutopayout)) != nil {
			m.responser.Error(c, http.StatusNotFound, AppError.ErrRecordNotFound)
			return
		}

		// Декодирую опциональные параметры
		switch obj.(*models.MerchantAutopayout).Service {
		case AppType.MerchantAutoPayoutWhitebit:
			p, err := m.pl.Whitebit.GetOptionParams(obj.(*models.MerchantAutopayout).Options)
			if err != nil {
				m.responser.Error(c, http.StatusUnprocessableEntity, AppError.ErrMerchantAutopatoutOptionalParams)
				return
			}

			// Делаю запрос на сервис мерчанта/автовыплаты
			b, err := m.pl.Whitebit.Ping(p)
			if err != nil {
				m.responser.Error(c, http.StatusInternalServerError, err)
				return
			}

			var resp map[string]interface{}

			if err := json.Unmarshal([]byte(b.([]byte)), &resp); err != nil {
				m.responser.Error(c, http.StatusInternalServerError, err)
				return
			}

			if resp["code"] != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":     AppError.ErrConnectionFailed.Error(),
					"meta_data": resp,
				})
				return
			}

			c.JSON(http.StatusOK, resp)
			return
		}
	}

	m.responser.Error(c, http.StatusInternalServerError, AppError.ErrFailedToInitializeStruct)
}
