package bills

import (
	"github.com/gefion-tech/tg-exchanger-server/internal/config"
	AppType "github.com/gefion-tech/tg-exchanger-server/internal/core/types"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/nsqstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/redisstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/utils"
	"github.com/gin-gonic/gin"
)

type ModBills struct {
	store db.SQLStoreI
	redis *redisstore.AppRedisDictionaries
	nsq   nsqstore.NsqI
	cfg   *config.Config

	responser utils.ResponserI
	logger    utils.LoggerI
}

type ModBillsI interface {
	GetBillHandler(c *gin.Context)
	DeleteBillHandler(c *gin.Context)
	GetAllBillsHandler(c *gin.Context)

	RejectBillHandler(c *gin.Context)
	CreateBillHandler(c *gin.Context)
}

func InitModBills(
	store db.SQLStoreI,
	redis *redisstore.AppRedisDictionaries,
	nsq nsqstore.NsqI,
	cfg *config.Config,
	responser utils.ResponserI,
	l utils.LoggerI,
) ModBillsI {
	return &ModBills{
		store: store,
		redis: redis,
		nsq:   nsq,
		cfg:   cfg,

		responser: responser,
		logger:    l,
	}
}

func (m *ModBills) modlog(err error) {
	m.logger.NewRecord(&models.LogRecord{
		Service: AppType.LogTypeServer,
		Module:  "BILL_HANDLER_MOD",
		Info:    err.Error(),
	})
}
