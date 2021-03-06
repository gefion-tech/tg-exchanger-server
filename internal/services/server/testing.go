package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gefion-tech/tg-exchanger-server/internal/config"
	"github.com/gefion-tech/tg-exchanger-server/internal/mocks"
	"github.com/gefion-tech/tg-exchanger-server/internal/plugins"
	mine_plugin "github.com/gefion-tech/tg-exchanger-server/internal/plugins/mine"
	whitebit_plugin "github.com/gefion-tech/tg-exchanger-server/internal/plugins/whitebit"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/mocksqlstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/nsqstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/redisstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/utils"
	"github.com/stretchr/testify/assert"
)

/*
	Функция возвращает сконфигурированный тестовый сервер
	К тестовому серверу подключается имитация sql хранилища
*/
func TestServer(t *testing.T) (*Server, *redisstore.AppRedisDictionaries, func(*redisstore.AppRedisDictionaries)) {
	t.Helper()

	config := config.InitTestConfig(t)
	assert.NotNil(t, config)

	// Создание redis хранилища для хранения данных о регистрации пользователя
	rRegistration, err := db.InitRedis(&config.Services.Redis, 1)
	assert.NoError(t, err)

	// Создание redis хранилища для хранения пользовательских сессий
	rAuth, err := db.InitRedis(&config.Services.Redis, 2)
	assert.NoError(t, err)

	AppRedis := &redisstore.AppRedisDictionaries{
		Registration: redisstore.InitRegistrationClient(rRegistration),
		Auth:         redisstore.InitAuthClient(rAuth),
	}

	// Инициализация соединения с NSQ
	producer, err := db.InitNSQ(&config.Services.NSQ)
	assert.NoError(t, err)

	logger := utils.InitLogger(mocksqlstore.Init().AdminPanel().Logs())
	plugins := plugins.InitAppPlugins(
		mine_plugin.InitMinePlugin(),
		whitebit_plugin.InitWhitebitPlugin(&config.Plugins),
	)

	return root(mocksqlstore.Init(), nsqstore.Init(producer), AppRedis, plugins, logger, config), AppRedis, func(appRedis *redisstore.AppRedisDictionaries) {
		appRedis.Registration.Clear()
		appRedis.Registration.Close()

		appRedis.Auth.Clear()
		appRedis.Auth.Close()
	}
}

/*
	Метод для быстрой проверки текста ошибки
*/
func TestGetErrorText(t *testing.T, recBody *bytes.Buffer) (string, error) {
	t.Helper()

	var body map[string]interface{}

	if err := json.NewDecoder(recBody).Decode(&body); err != nil {
		return "", err
	}

	return body["error"].(string), nil
}

/*
	==========================================================================================
	ФУНКЦИИ СОЗДАНИЯ ТЕСТОВЫХ ОБЪЕКТОВ
	==========================================================================================
*/

func TestMerchantAutopayout(t *testing.T, s *Server, tokens map[string]interface{}) error {
	t.Helper()

	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(mocks.MerchantAutopayout); err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/v1/admin/merchant-autopayout", b)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens["access_token"]))
	s.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		return errors.New("не удалось создать тестовую запись")
	}

	var body map[string]interface{}

	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		return err
	}

	return nil
}

func TestLogRecord(t *testing.T, s *Server) error {
	t.Helper()

	// Кодирую тело запроса
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(mocks.LOG_RECORD__ADMIN); err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/v1/log", b)
	if err != nil {
		return err
	}

	s.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		return errors.New("не удалось создать тестовую лог запись")
	}

	var body map[string]interface{}

	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		return err
	}

	return nil
}

func TestNotification854(t *testing.T, s *Server, tokens map[string]interface{}) error {
	t.Helper()

	// Кодирую тело запроса
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(mocks.ADMIN_NOTIFICATION_854); err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/v1/admin/notification", b)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens["access_token"]))
	s.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		return errors.New("не удалось создать тествое уведомление")
	}

	return nil
}

func TestBotMessage(t *testing.T, s *Server, tokens map[string]interface{}) error {
	t.Helper()

	// Кодирую тело запроса
	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(mocks.BOT_MESSAGE_REQ); err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/v1/admin/message", b)
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens["access_token"]))
	s.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		return errors.New("не удалось создать тествое сообщение")
	}

	return nil
}

func TestManager(t *testing.T, s *Server) (map[string]interface{}, error) {
	t.Helper()

	// Регистраци нового пользователя бота
	if err := TestBotUser(t, s); err != nil {
		return nil, err
	}

	/* Регистраци менеджера -> ШАГ 1 */

	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(map[string]interface{}{
		"username": mocks.MANAGER_IN_ADMIN_REQ["username"],
		"password": mocks.MANAGER_IN_ADMIN_REQ["password"],
		"testing":  true,
	}); err != nil {
		return nil, err
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/v1/admin/registration/code", b)
	if err != nil {
		return nil, err
	}
	s.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return nil, fmt.Errorf("registration step 1 | Status code %d", rec.Code)
	}

	/* Регистраци менеджера -> ШАГ 2 */

	// Кодирую тело запроса
	b2 := &bytes.Buffer{}
	json.NewEncoder(b2).Encode(map[string]interface{}{"code": 100000})

	rec2 := httptest.NewRecorder()
	req2, err := http.NewRequest(http.MethodPost, "/api/v1/admin/registration", b2)
	if err != nil {
		return nil, err
	}
	s.Router.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusCreated {
		return nil, fmt.Errorf("registration step 2 | Status code %d", rec2.Code)
	}

	/* Авторизация */

	b3 := &bytes.Buffer{}
	json.NewEncoder(b3).Encode(mocks.MANAGER_IN_ADMIN_REQ)

	rec3 := httptest.NewRecorder()
	req3, err := http.NewRequest(http.MethodPost, "/api/v1/admin/auth", b3)
	if err != nil {
		return nil, err
	}
	s.Router.ServeHTTP(rec3, req3)

	tokens := map[string]interface{}{}

	if err := json.NewDecoder(rec3.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

/*
	Функция для быстрой регистраци пользователя в боте
*/
func TestBotUser(t *testing.T, s *Server) error {
	t.Helper()

	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(mocks.USER_IN_BOT_REGISTRATION_REQ); err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/v1/bot/user/registration", b)
	if err != nil {
		return err
	}

	s.Router.ServeHTTP(rec, req)
	return nil
}

func TestExchanger(t *testing.T, s *Server, tokens map[string]interface{}) error {
	t.Helper()

	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(mocks.ADMIN_EXCHANGER); err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/v1/admin/exchanger", b)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens["access_token"]))
	if err != nil {
		return err
	}

	s.Router.ServeHTTP(rec, req)

	return nil
}

func TestUserBill(t *testing.T, s *Server, tokens map[string]interface{}) error {
	t.Helper()

	b := &bytes.Buffer{}
	if err := json.NewEncoder(b).Encode(mocks.USER_BILL_REQ); err != nil {
		return err
	}

	rec := httptest.NewRecorder()
	req, err := http.NewRequest(http.MethodPost, "/api/v1/admin/bill", b)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokens["access_token"]))
	if err != nil {
		return err
	}

	s.Router.ServeHTTP(rec, req)
	return nil
}
