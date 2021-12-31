package notification

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gefion-tech/tg-exchanger-server/internal/app/errors"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gin-gonic/gin"
)

/*
	@Method GET
	@Path admin/notifications?page=1&limit=15
	@Type PRIVATE
	@Documentation

	Получение лимитированного объема записей из таблицы `notifications`

	# TESTED

*/
func (m *ModNotification) GetNotificationsSelectionHandler(c *gin.Context) {
	m.responser.SelectionResponse(c, m.store.AdminPanel().Notification())
}

/*
	@Method PUT
	@Path admin/notification
	@Type PRIVATE
	@Documentation

	Обновить поле status записи в таблице `notifications`

	# TESTED
*/
func (m *ModNotification) UpdateNotificationStatusHandler(c *gin.Context) {
	r := &models.Notification{}
	if err := c.ShouldBindJSON(r); err != nil {
		m.responser.Error(c, http.StatusUnprocessableEntity, errors.ErrInvalidBody)
		return
	}

	m.responser.Error(c, http.StatusUnprocessableEntity,
		r.NotificationStatusValidation(),
		r.NotificationTypeValidation(),
	)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		m.responser.Error(c, http.StatusUnprocessableEntity, err)
		return
	}

	r.ID = id

	m.responser.RecordResponse(c, r, m.store.AdminPanel().Notification().UpdateStatus(r))
}

/*
	@Method DELETE
	@Path admin/notification
	@Type PRIVATE
	@Documentation

	Удалить запись в таблице `notifications`

	# TESTED
*/
func (m *ModNotification) DeleteNotificationHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		m.responser.Error(c, http.StatusUnprocessableEntity, err)
		return
	}

	r := &models.Notification{ID: id}
	m.responser.RecordResponse(c, r, m.store.AdminPanel().Notification().Delete(r))
}

func newSupportReqNotify(uArr []*models.User, i int, n *models.Notification) map[string]interface{} {
	return map[string]interface{}{
		"to": map[string]interface{}{
			"chat_id":  uArr[i].ChatID,
			"username": uArr[i].Username,
		},
		"message": map[string]interface{}{
			"type": "confirmation_req",
			"text": fmt.Sprintf("🔵 Запрос тех. поддержки 🔵\n\n*Пользователь*: @%s", n.User.Username),
		},
		"created_at": time.Now().UTC().Format("2006-01-02T15:04:05.00000000"),
	}
}

func newVefificationNotify(uArr []*models.User, i int, n *models.Notification) map[string]interface{} {
	return map[string]interface{}{
		"to": map[string]interface{}{
			"chat_id":  uArr[i].ChatID,
			"username": uArr[i].Username,
		},
		"message": map[string]interface{}{
			"type": "confirmation_req",
			"text": fmt.Sprintf("🟢 Новая заявка 🟢\n\n*Пользователь*: @%s", n.User.Username),
		},
		"created_at": time.Now().UTC().Format("2006-01-02T15:04:05.00000000"),
	}
}

func newActionCancelNotify(uArr []*models.User, i int, n *models.Notification) map[string]interface{} {
	return map[string]interface{}{
		"to": map[string]interface{}{
			"chat_id":  uArr[i].ChatID,
			"username": uArr[i].Username,
		},
		"message": map[string]interface{}{
			"type": "skip_operation",
			"text": fmt.Sprintf("🔴 Отмена операции 🔴\n\n*Пользователь*: @%s", n.User.Username),
		},
		"created_at": time.Now().UTC().Format("2006-01-02T15:04:05.00000000"),
	}
}
