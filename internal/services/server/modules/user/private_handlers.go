package user

import (
	"net/http"

	AppError "github.com/gefion-tech/tg-exchanger-server/internal/core/errors"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/guard"
	"github.com/gin-gonic/gin"
)

/*
	@Method POST
	@Path admin/logout
	@Type PRIVATE
	@Documentation https://github.com/exchanger-bot/docs/blob/main/admin/account/account_api.md#logout

	При валидных данных токена в Redis удаляется
	токен используемый в текущей сессии.

	# TESTED
*/
func (m *ModUsers) LogoutHandler(c *gin.Context) {
	// Извлекаю метаданные JWT
	ctxToken := c.Request.Context().Value(guard.CtxKeyToken).(*models.AccessDetails)

	// Удаляю токен
	deleted, err := m.redis.Auth.DeleteAuth(ctxToken.AccessUuid)
	if err != nil || deleted == 0 {
		m.responser.Error(c, http.StatusUnauthorized, AppError.ErrUnauthorized)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
