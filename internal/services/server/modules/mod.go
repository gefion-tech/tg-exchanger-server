package modules

import (
	"github.com/gefion-tech/tg-exchanger-server/internal/config"
	AppType "github.com/gefion-tech/tg-exchanger-server/internal/core/types"
	"github.com/gefion-tech/tg-exchanger-server/internal/plugins"
	mine_plugin "github.com/gefion-tech/tg-exchanger-server/internal/plugins/mine"
	whitebit_plugin "github.com/gefion-tech/tg-exchanger-server/internal/plugins/whitebit"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/nsqstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/redisstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/guard"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/middleware"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/modules/bills"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/modules/directions"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/modules/exchanger"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/modules/logs"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/modules/ma"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/modules/message"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/modules/notification"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/server/modules/user"
	"github.com/gefion-tech/tg-exchanger-server/internal/utils"
	"github.com/gin-gonic/gin"
)

type ServerModules struct {
	exMod        exchanger.ModExchangerI
	notifyMod    notification.ModNotificationI
	userMod      user.ModUsersI
	msgMod       message.ModMessageI
	billsMod     bills.ModBillsI
	logsMod      logs.ModLogsI
	maMod        ma.ModMerchantAutoPayoutI
	directionMod directions.ModDirectionsI
}

type ServerModulesI interface {
	ModulesConfigure(router *gin.RouterGroup, g guard.GuardI, mdl middleware.MiddlewareI)
}

func InitServerModules(
	store db.SQLStoreI,
	redis *redisstore.AppRedisDictionaries,
	nsq nsqstore.NsqI,
	cfg *config.Config,
	logger utils.LoggerI,
	responser utils.ResponserI,
) ServerModulesI {
	return &ServerModules{
		exMod: exchanger.InitModExchanger(
			store,
			redis,
			nsq,
			cfg,
			responser,
			logger,
		),

		notifyMod: notification.InitModNotification(
			store,
			redis,
			nsq,
			cfg,
			logger,
			responser,
		),

		msgMod: message.InitModMessage(
			store,
			redis,
			nsq,
			cfg,
			responser,
			logger,
		),

		userMod: user.InitModUsers(
			store,
			redis,
			nsq,
			cfg,
			responser,
			logger,
		),

		billsMod: bills.InitModBills(
			store,
			redis,
			nsq,
			cfg,
			responser,
			logger,
		),

		logsMod: logs.InitModLogs(
			store.AdminPanel().Logs(),
			cfg,
			responser,
		),

		maMod: ma.InitModMerchantAutoPayout(
			store.AdminPanel(),
			redis,
			nsq,
			cfg,
			plugins.InitAppPlugins(
				mine_plugin.InitMinePlugin(),
				whitebit_plugin.InitWhitebitPlugin(&cfg.Plugins),
			),
			responser,
			logger,
		),

		directionMod: directions.InitModDirections(store.AdminPanel(), cfg, responser),
	}
}

func (m *ServerModules) ModulesConfigure(router *gin.RouterGroup, g guard.GuardI, mdl middleware.MiddlewareI) {
	// base
	{
		router.POST(
			"/bot/user/registration",
			m.userMod.UserInBotRegistrationHandler,
		)
	}

	// bot bill
	{
		router.GET(
			"/bot/user/bill/:id",
			m.billsMod.GetBillHandler,
		)
		router.GET(
			"/bot/user/:chat_id/bills",
			m.billsMod.GetAllBillsHandler,
		)
		router.DELETE(
			"/bot/user/:chat_id/bill/:id",
			m.billsMod.DeleteBillHandler,
		)
	}

	// bill
	{
		router.POST(
			"/admin/bill",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceBill, AppType.ResourceCreate),
			m.billsMod.CreateBillHandler,
		)
		router.POST(
			"/admin/bill/reject",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceBill, AppType.ResourceReject),
			m.billsMod.RejectBillHandler,
		)
	}

	// registration|auth
	{
		router.POST(
			"/admin/registration/code",
			m.userMod.UserGenerateCodeHandler,
		)
		router.POST(
			"/admin/registration",
			m.userMod.UserInAdminRegistrationHandler,
		)
		router.POST(
			"/admin/auth",
			m.userMod.UserInAdminAuthHandler,
		)
		router.POST(
			"/admin/token/refresh",
			m.userMod.UserRefreshToken,
		)
		router.POST(
			"/admin/logout",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.userMod.LogoutHandler,
		)
	}

	// message
	{
		router.POST(
			"/admin/message",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceMessage, AppType.ResourceCreate),
			m.msgMod.CreateNewMessageHandler,
		)
		router.PUT(
			"/admin/message/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceMessage, AppType.ResourceUpdate),
			m.msgMod.UpdateBotMessageHandler,
		)
		router.DELETE(
			"/admin/message/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceMessage, AppType.ResourceDelete),
			m.msgMod.DeleteBotMessageHandler,
		)
		router.GET(
			"/admin/message/:connector",
			m.msgMod.GetMessageHandler,
		)
		router.GET(
			"/admin/messages",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.msgMod.GetMessagesSelectionHandler,
		)
	}

	// notification
	{
		router.POST(
			"/admin/notification",
			m.notifyMod.CreateNotificationHandler,
		)
		router.PUT(
			"/admin/notification/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.notifyMod.UpdateNotificationStatusHandler,
		)
		router.DELETE(
			"/admin/notification/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceNotify, AppType.ResourceDelete),
			m.notifyMod.DeleteNotificationHandler,
		)
		router.GET(
			"/admin/notifications",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.notifyMod.GetNotificationsSelectionHandler,
		)
		router.GET(
			"/admin/notifications/check",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.notifyMod.NewNotificationsCheckHandler,
		)
	}

	// exchanger
	{
		router.POST(
			"/admin/exchanger",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceExchange, AppType.ResourceCreate),
			m.exMod.CreateExchangerHandler,
		)
		router.PUT(
			"/admin/exchanger/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceExchange, AppType.ResourceUpdate),
			m.exMod.UpdateExchangerHandler,
		)
		router.DELETE(
			"/admin/exchanger/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceExchange, AppType.ResourceDelete),
			m.exMod.DeleteExchangerHandler,
		)
		router.GET(
			"/admin/exchanger/:name",
			m.exMod.GetExchangerByNameHandler,
		)
		router.GET(
			"/admin/exchanger/document",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.exMod.GetExchangerDocumentHandler,
		)
		router.GET(
			"/admin/exchangers",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.exMod.GetExchangersSelectionHandler,
		)
	}

	// merchant/autopayout
	{
		router.POST(
			"admin/merchant-autopayout",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceMerchantAutopayout, AppType.ResourceCreate),
			m.maMod.CreateMerchantAutopayoutHandler,
		)
		router.PUT(
			"admin/merchant-autopayout/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceMerchantAutopayout, AppType.ResourceUpdate),
			m.maMod.UpdateMerchantAutopayoutHandler,
		)
		router.DELETE(
			"admin/merchant-autopayout/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.Logger(AppType.ResourceMerchantAutopayout, AppType.ResourceDelete),
			m.maMod.DeleteMerchantAutopayoutHandler,
		)
		router.GET(
			"admin/merchant-autopayout/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.maMod.GetMerchantAutopayoutHandler,
		)
		router.GET(
			"admin/merchant-autopayout/ping/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.maMod.PingMerchantAutopayoutHandler,
		)
		router.GET(
			"admin/merchant-autopayout/history/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.maMod.GetHistoryMerchantAutopayoutHandler,
		)
		router.GET(
			"admin/merchant-autopayout/balance/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.maMod.GetBalanceMerchantAutopayoutHandler,
		)
		router.GET(
			"admin/merchant-autopayout/all",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.maMod.GetMerchantAutopayoutSelectionHandler,
		)
		router.POST(
			"admin/merchant-autopayout/:service/new-adress",
			m.maMod.CreateNewAdressHandler,
		)
	}

	// directions
	{
		router.POST(
			"/admin/direction",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.directionMod.CreateDirectionHandler,
		)
		router.PUT(
			"/admin/direction/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.directionMod.UpdateDirectionHandler,
		)
		router.GET(
			"/admin/direction/:id",
			m.directionMod.GetDirectionHandler,
		)
		router.DELETE(
			"/admin/direction/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.directionMod.DeleteDirectionHandler,
		)

		router.GET(
			"/admin/directions",
			g.AuthTokenValidation(),
			g.IsAuth(),
			m.directionMod.DirectionSelectionHandler,
		)

		// Merchant and autopayout for direction
		{
			router.POST(
				"/admin/direction/merchant-autopayout",
				g.AuthTokenValidation(),
				g.IsAuth(),
				m.directionMod.DeleteDirectionMaHandler,
			)
			router.PUT(
				"/admin/direction/merchant-autopayout/:id",
				g.AuthTokenValidation(),
				g.IsAuth(),
				m.directionMod.UpdateDirectionMaHandler,
			)
			router.GET(
				"/admin/direction/merchant-autopayout/:id",
				g.AuthTokenValidation(),
				g.IsAuth(),
				m.directionMod.GetDirectionMaHandler,
			)
			router.DELETE(
				"/admin/direction/merchant-autopayout/:id",
				g.AuthTokenValidation(),
				g.IsAuth(),
				m.directionMod.DeleteDirectionMaHandler,
			)
			router.GET(
				"/admin/direction/merchant-autopayout/all",
				g.AuthTokenValidation(),
				g.IsAuth(),
				m.directionMod.DirectionMaSelectionHandler,
			)
		}
	}

	// log
	{
		router.POST(
			"/log",
			m.logsMod.CreateLogRecordHandler,
		)
		router.DELETE(
			"/log/:id",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.IsAdmin(),
			m.logsMod.DeleteLogRecordHandler,
		)
		router.GET(
			"/logs",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.IsAdmin(),
			m.logsMod.GetLogRecordsSelectionHandler,
		)
		router.DELETE(
			"/logs",
			g.AuthTokenValidation(),
			g.IsAuth(),
			g.IsAdmin(),
			m.logsMod.DeleteLogRecordsSelectionHandler,
		)
	}
}
