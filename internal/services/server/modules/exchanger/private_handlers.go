package exchanger

import (
	"encoding/json"
	"net/http"
	"reflect"

	AppError "github.com/gefion-tech/tg-exchanger-server/internal/core/errors"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gin-gonic/gin"
)

/*
	@Method POST
	@Path admin/exchangers
	@Type PRIVATE
	@Documentation

	Получение лимитированного объема записей из таблицы `exchangers`

	# TESTED
*/
func (m *ModExchanger) GetExchangersSelectionHandler(c *gin.Context) {
	m.responser.SelectionResponse(c,
		m.store.AdminPanel().Exchanger(),
		&models.ExchangerSelection{},
	)
}

/*
	@Method DELETE
	@Path admin/exchanger/:id
	@Type PRIVATE
	@Documentation

	Удалить запись в таблице `exchangers`

	# TESTED
*/
func (m *ModExchanger) DeleteExchangerHandler(c *gin.Context) {
	if obj := m.responser.RecordHandler(c, &models.Exchanger{}); obj != nil {
		// Проверю, удалось ли записать структуру или была поймана ошибка
		if reflect.TypeOf(obj) != reflect.TypeOf(&models.Exchanger{}) {
			return
		}

		m.responser.DeleteRecordResponse(c, m.store.AdminPanel().Exchanger(), obj)
	}

	m.responser.Error(c, http.StatusInternalServerError, AppError.ErrFailedToInitializeStruct)
}

/*
	@Method PUT
	@Path admin/exchanger/:id
	@Type PRIVATE
	@Documentation

	Обновить запись в таблице `exchangers`

	# TESTED
*/
func (m *ModExchanger) UpdateExchangerHandler(c *gin.Context) {
	// Декодирование
	r := &models.Exchanger{}
	if err := c.ShouldBindJSON(r); err != nil {
		m.responser.Error(c, http.StatusUnprocessableEntity, AppError.ErrInvalidBody)
		return
	}

	if obj := m.responser.RecordHandler(c, r, r.ExchangerUpdateValidation()); obj != nil {
		// Проверю, удалось ли записать структуру или была поймана ошибка
		if reflect.TypeOf(obj) != reflect.TypeOf(&models.Exchanger{}) {
			return
		}

		// Отправляю в очередь команду с новой ссылкой
		// Это необходимо для корректной работы микросервиса
		// обработки котировок
		payload, err := json.Marshal(map[string]interface{}{
			"Name": "update_url",
			"MetaData": map[string]interface{}{
				"NewUrl": r.UrlToParse,
			},
		})
		if err != nil {
			m.responser.Error(c, http.StatusInternalServerError, err)
		}

		if err := m.nsq.Publish("quotes-ms", payload); err != nil {
			m.responser.Error(c, http.StatusInternalServerError, err)
		}

		m.responser.UpdateRecordResponse(c, m.store.AdminPanel().Exchanger(), obj)
		return
	}

	m.responser.Error(c, http.StatusInternalServerError, AppError.ErrFailedToInitializeStruct)
}

/*
	@Method POST
	@Path admin/exchanger
	@Type PRIVATE
	@Documentation

	Создать запись в таблице `exchangers`

	# TESTED
*/
func (m *ModExchanger) CreateExchangerHandler(c *gin.Context) {
	// Декодирование
	r := &models.Exchanger{}
	if err := c.ShouldBindJSON(r); err != nil {
		m.responser.Error(c, http.StatusUnprocessableEntity, AppError.ErrInvalidBody)
		return
	}

	if obj := m.responser.RecordHandler(c, r, r.ExchangerCreateValidation()); obj != nil {
		// Проверю, удалось ли записать структуру или была поймана ошибка
		if reflect.TypeOf(obj) != reflect.TypeOf(&models.Exchanger{}) {
			return
		}

		m.responser.CreateRecordResponse(c, m.store.AdminPanel().Exchanger(), obj)
		return
	}

	m.responser.Error(c, http.StatusInternalServerError, AppError.ErrFailedToInitializeStruct)
}
