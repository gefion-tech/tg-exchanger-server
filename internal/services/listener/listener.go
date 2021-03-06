package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/gefion-tech/tg-exchanger-server/internal/config"
	AppType "github.com/gefion-tech/tg-exchanger-server/internal/core/types"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gefion-tech/tg-exchanger-server/internal/plugins"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db"
	"github.com/gefion-tech/tg-exchanger-server/internal/services/db/nsqstore"
	"github.com/gefion-tech/tg-exchanger-server/internal/utils"
	"golang.org/x/sync/errgroup"
)

type Listener struct {
	store  db.SQLStoreI
	nsq    nsqstore.NsqI
	plugin *plugins.AppPlugins
	logger utils.LoggerI
}

type ListenerI interface {
	Listen(ctx context.Context, cfg *config.ListenerConfig) error
}

func InitListener(s db.SQLStoreI, q nsqstore.NsqI, p *plugins.AppPlugins, l utils.LoggerI) ListenerI {
	return &Listener{
		store:  s,
		nsq:    q,
		plugin: p,
		logger: l,
	}
}

func (listener *Listener) Listen(ctx context.Context, cfg *config.ListenerConfig) error {
	for {
		t := time.NewTimer(time.Duration(cfg.Interval) * time.Second)

		// Получение актуального списка аккаунтов
		state, err := listener.snapshot(ctx, cfg)
		if err != nil {
			return err
		}

		// Канал для массива историй транзакций со всех аккаунтов на Whitebit
		cWhitebitHistoryArr := make(chan []*models.WhitebitHistory)
		// Канал для массива всех заявок сохраненных в БД
		cExchangeRequestsArr := make(chan []*models.ExchangeRequest)

		// Массив транзакций со всех аккаунтов на Whitebit
		var whitebitHistoryArr []*models.WhitebitHistory
		// Массив всех заявок из БД
		var exchangeRequestsArr []*models.ExchangeRequest

		{
			errs, _ := errgroup.WithContext(ctx)

			// Получение списка новых заявок и заявок по которым
			// должны отработать автовыплаты
			errs.Go(func() error {
				defer close(cExchangeRequestsArr)

				arr, err := listener.store.AdminPanel().ExchangeRequest().GetAllByStatus(
					AppType.ExchangeRequestNew,                  // новые заявки
					AppType.ExchangeRequestPaid,                 // заявки по которым должны отработать автовыплаты
					AppType.ExchangeRequestAwaitingConfirmation, // заявки по которым отработали автовыплаты, ожидают подтверждения от биржи
				)
				if err != nil {
					return err
				}

				cExchangeRequestsArr <- arr
				return nil
			})

			// Получаю истории транзакций всех whitebit аккаутов
			errs.Go(func() error {
				defer close(cWhitebitHistoryArr)
				arr := []*models.WhitebitHistory{}

				for _, merchant := range state.Merchants.Whitebit {
					history, err := listener.checker(merchant)
					if err != nil {
						return err
					}

					arr = append(arr, history)
				}

				cWhitebitHistoryArr <- arr
				return nil
			})

			whitebitHistoryArr = <-cWhitebitHistoryArr
			exchangeRequestsArr = <-cExchangeRequestsArr

			if errs.Wait() != nil {
				fmt.Println(errs.Wait())
			}
		}

		// Анализ истории транзакций всех аккаунтов
		{
			errs, _ := errgroup.WithContext(ctx)

			// Анализ истории всех транзакций со всех аккаунтов на whitebit
			errs.Go(func() error {

				// Массив заявок по которым должны отработать автовыплаты
				var forAutopayout []*models.ExchangeRequest

				// rHistory -> Запись из истории транзакций
				// rRequest -> Запись в таблице заявок
				for _, account := range whitebitHistoryArr {
					for _, rHistory := range account.Records {
						for _, rRequest := range exchangeRequestsArr {
							// Если заявка ожидает автовыплаты
							if rRequest.Status == AppType.ExchangeRequestPaid {
								if len(forAutopayout) > 0 {
									for _, alreadyAdd := range forAutopayout {
										if alreadyAdd.ID != rRequest.ID {
											forAutopayout = append(forAutopayout, rRequest)
										}
									}
								} else {
									forAutopayout = append(forAutopayout, rRequest)
								}
								continue
							}

							switch rHistory.Method {
							case 1: // Событие получения средств
								listener.handleWhitebitDepositAction(rHistory, rRequest)
								continue
							case 2: // Событие вывода средств
								if rHistory.Status == 3 || rHistory.Status == 7 {
									listener.handleWhitebitWithdrawAction(rHistory, rRequest)
									continue
								}
							default:
								continue
							}
						}
					}
				}

				time.Sleep(time.Duration(1 * time.Second))

				// Работа автовыплаты
				for _, rRequest := range forAutopayout {
					for _, account := range state.Merchants.Whitebit {
						b, err := listener.plugin.Whitebit.AutoPayout().Payout(account, map[string]interface{}{
							"ticker":   rRequest.ExchangeTo,
							"amount":   fmt.Sprintf("%f", rRequest.ExpectedAmount),
							"address":  rRequest.ClientAddress,
							"uniqueId": strconv.Itoa(rRequest.ID),
							"network":  "TRC20",
						})
						if err != nil {
							fmt.Println(err)
						}

						var body interface{}
						if err := json.Unmarshal(b.([]byte), &body); err != nil {
							fmt.Println(err)
							break
						}

						// Если получили ошибку и деньги не отправились
						if reflect.TypeOf(body) == reflect.TypeOf(map[string]interface{}{}) {
							resp := body.(map[string]interface{})
							utils.SetSuccessStep(AppType.SprintfStep("Payout done with status %v", resp["code"]))
							fmt.Println(resp["errors"])
							continue
						}

						// Если деньги ушли
						rRequest.Status = AppType.ExchangeRequestAwaitingConfirmation
						if err := listener.store.AdminPanel().ExchangeRequest().Update(rRequest); err != nil {
							fmt.Println(err)
						}

					}
				}

				return nil
			})

			if errs.Wait() != nil {
				fmt.Println(errs.Wait())
			}
		}

		<-t.C
	}
}

func (listener *Listener) checker(p *models.WhitebitOptionParams) (*models.WhitebitHistory, error) {
	time.Sleep(time.Duration(1 * time.Second))

	b, err := listener.plugin.Whitebit.History(p, AppType.BaseWhitebitGetHistoryBody)
	if err != nil {
		// TODO: Писать лог что не удалось установить соединение с этим аккаунтом
		return nil, err
	}

	var history models.WhitebitHistory
	if err := json.Unmarshal(b.([]byte), &history); err != nil {
		return nil, err
	}

	return &history, nil
}
