package private

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gefion-tech/tg-exchanger-server/internal/app/errors"
	"github.com/gefion-tech/tg-exchanger-server/internal/app/static"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/nsqstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/tools"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

/*
	@Method POST
	@Path admin/notification
	@Type PUBLIC
	@Documentation

	При валидных данных в БД создается запись о новом уведомлении.

	# TESTED
*/
func (pr *PrivateRoutes) createNotification(c *gin.Context) {
	req := &models.Notification{}
	if err := c.ShouldBindJSON(req); err != nil {
		tools.ServErr(c, http.StatusUnprocessableEntity, errors.ErrInvalidBody)
		return
	}

	// Валидация типа уведомления
	if err := req.NotificationTypeValidation(); err != nil {
		tools.ServErr(c, http.StatusUnprocessableEntity, err)
		return
	}

	// Получаю всех менеджеров из БД
	uArr, err := pr.store.User().GetAllManagers()
	if err != nil {
		fmt.Println(err)
	}

	switch req.Type {
	case static.NTF__T__VERIFICATION:
		// Запись уведомления в очередь для всех менеджеров
		for i := 0; i < len(uArr); i++ {
			payload, err := json.Marshal(map[string]interface{}{
				"to": map[string]interface{}{
					"chat_id":  uArr[i].ChatID,
					"username": uArr[i].Username,
				},
				"message": map[string]interface{}{
					"type": "confirmation_req",
					"text": fmt.Sprintf("🟢 Новая заявка 🟢\n\n*Пользователь*: @%s", req.User.Username),
				},
				"created_at": time.Now().UTC().Format("2006-01-02T15:04:05.00000000"),
			})
			if err != nil {
				fmt.Println(err)
			}

			if err := pr.nsq.Publish(nsqstore.TOPIC__MESSAGES, payload); err != nil {
				fmt.Println(err)
			}
		}
	}

	n, err := pr.store.Manager().Notification().Create(req)
	switch err {
	case nil:
		c.JSON(http.StatusCreated, n)
		return
	default:
		tools.ServErr(c, http.StatusUnprocessableEntity, err)
		return
	}

}

/*
	@Method GET
	@Path admin/notifications?page=1&limit=15
	@Type PRIVATE
	@Documentation

	При валидных данных из БД достаюется нужные записи в нужном количестве

	# TESTED

*/
func (pr *PrivateRoutes) getAllNotifications(c *gin.Context) {
	errs, _ := errgroup.WithContext(c)

	cArrN := make(chan []*models.Notification)
	cCount := make(chan *int)

	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		tools.ServErr(c, http.StatusUnprocessableEntity, err)
	}

	limit, err := strconv.Atoi(c.DefaultQuery("limit", "15"))
	if err != nil {
		tools.ServErr(c, http.StatusUnprocessableEntity, err)
		return
	}

	// Достаю из БД запрашиваемые записи
	errs.Go(func() error {
		defer close(cArrN)
		arrN, err := pr.store.Manager().Notification().GetSlice(page * limit)
		if err != nil {
			return err
		}

		cArrN <- arrN
		return nil
	})

	// Подсчет кол-ва уведомлений в таблице
	errs.Go(func() error {
		defer close(cCount)
		c, err := pr.store.Manager().Notification().Count()
		if err != nil {
			return err
		}

		cCount <- &c
		return nil
	})

	arrN := <-cArrN
	count := <-cCount

	if arrN == nil || count == nil {
		tools.ServErr(c, http.StatusInternalServerError, errs.Wait())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"limit":        limit,
		"current_page": page,
		"last_page":    math.Ceil(float64(*count) / float64(limit)),
		"data":         arrN[(page-1)*limit : tools.UpperThreshold(page, limit, *count)],
	})
}

/*
	@Method PUT
	@Path admin/notification
	@Type PRIVATE
	@Documentation

	Обновить статус уведомления.

	# TESTED
*/
func (pr *PrivateRoutes) updateNotificationStatus(c *gin.Context) {
	req := &models.Notification{}
	if err := c.ShouldBindJSON(req); err != nil {
		tools.ServErr(c, http.StatusUnprocessableEntity, errors.ErrInvalidBody)
		return
	}

	if err := req.NotificationTypeValidation(); err != nil {
		tools.ServErr(c, http.StatusUnprocessableEntity, err)
		return
	}

	if err := req.NotificationStatusValidation(); err != nil {
		tools.ServErr(c, http.StatusUnprocessableEntity, err)
		return
	}

	n, err := pr.store.Manager().Notification().UpdateStatus(req)
	switch err {
	case nil:
		c.JSON(http.StatusOK, n)
		return
	case sql.ErrNoRows:
		tools.ServErr(c, http.StatusNotFound, errors.ErrRecordNotFound)
		return
	default:
		tools.ServErr(c, http.StatusNotFound, err)
		return
	}
}

/*
	@Method DELETE
	@Path admin/notification
	@Type PRIVATE
	@Documentation

	Удалить конкретное уведомление
*/
func (pr *PrivateRoutes) deleteNotification(c *gin.Context) {
	req := &models.Notification{}
	if err := c.ShouldBindJSON(req); err != nil {
		tools.ServErr(c, http.StatusUnprocessableEntity, errors.ErrInvalidBody)
		return
	}

	n, err := pr.store.Manager().Notification().Delete(req)
	switch err {
	case nil:
		c.JSON(http.StatusOK, n)
		return
	case sql.ErrNoRows:
		tools.ServErr(c, http.StatusNotFound, errors.ErrRecordNotFound)
		return
	default:
		tools.ServErr(c, http.StatusNotFound, err)
		return
	}

}
