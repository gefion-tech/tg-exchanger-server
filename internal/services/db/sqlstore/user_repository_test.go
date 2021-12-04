package sqlstore_test

import (
	"testing"

	"github.com/gefion-tech/tg-exchanger-server/internal/app/config"
	"github.com/gefion-tech/tg-exchanger-server/internal/mocks"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/sqlstore"
	"github.com/stretchr/testify/assert"
)

func Test_SQL_UserRepository(t *testing.T) {
	config := config.InitTestConfig(t)

	db, teardown := db.TestDB(t, &config.DB)
	defer teardown("users")

	// Вызываю создание хранилища
	s := sqlstore.Init(db)

	// Регистрация человека как пользователя бота
	u, err := s.User().Create(&mocks.USER_IN_BOT_REGISTRATION_REQUEST)
	assert.NoError(t, err)
	assert.NotNil(t, u)

	// Регистрация человека как менеджера
	m, err := s.User().RegisterAsManager(u)
	assert.NoError(t, err)
	assert.NotNil(t, m)

	// Поиск пользователя по его username
	uUsername, err := s.User().FindByUsername(u.Username)
	assert.NoError(t, err)
	assert.NotNil(t, uUsername)
}
