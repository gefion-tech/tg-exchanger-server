package mocks

import "github.com/gefion-tech/tg-exchanger-server/internal/models"

var USER_IN_BOT_REGISTRATION_REQUEST = models.UserFromBotRequest{
	ChatID:   3673563,
	Username: "I0HuKc",
}

var USER_IN_BOT_REGISTRATION_REQ = map[string]interface{}{
	"chat_id":  3673563,
	"username": "I0HuKc",
}

var MANAGER_IN_ADMIN_REQ = map[string]interface{}{
	"username": USER_IN_BOT_REGISTRATION_REQ["username"],
	"password": "4tfgefhey75uh",
}

var USER_BILL_REQ = map[string]interface{}{
	"chat_id": USER_IN_BOT_REGISTRATION_REQ["chat_id"],
	"bill":    "5559494130410827",
}
