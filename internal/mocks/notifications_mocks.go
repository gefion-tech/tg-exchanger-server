package mocks

import AppTypes "github.com/gefion-tech/tg-exchanger-server/internal/core/types"

var ADMIN_NOTIFICATION_854 = map[string]interface{}{
	"id":     1,
	"status": 1,
	"type":   AppTypes.NotifyTypeVerification,
	"meta_data": map[string]interface{}{
		"card_verification": map[string]interface{}{
			"code":      245335,
			"user_card": "5559494130410854",
			"img_path":  "tmp/some_path.png",
		},
	},
	"user": USER_IN_BOT_REGISTRATION_REQ,
}
