package models

import (
	_errors "errors"

	"github.com/gefion-tech/tg-exchanger-server/internal/app/static"
	validation "github.com/go-ozzo/ozzo-validation"
)

type Notification struct {
	ID       uint `json:"id"`
	Type     int  `json:"type" binding:"required"`
	Status   int  `json:"status"`
	MetaData struct {
		Code     *int    `json:"code"`
		UserCard *string `json:"user_card"`
		ImgPath  *string `json:"img_path"`
	} `json:"meta_data"`
	User struct {
		ChatID   int64  `json:"chat_id" binding:"required"`
		Username string `json:"username" binding:"required"`
	} `json:"user"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (n *Notification) NotificationTypeValidation() error {
	return validation.ValidateStruct(
		n,
		validation.Field(&n.Type, validation.By(nTypeValidation(n.Type))),
	)
}

func (n *Notification) NotificationStatusValidation() error {
	return validation.ValidateStruct(
		n,
		validation.Field(&n.Status, validation.By(nStatusValidation(n.Status))),
	)
}

func nTypeValidation(s int) validation.RuleFunc {
	return func(value interface{}) error {
		permitted := []int{static.NTF__T__VERIFICATION, static.NTF__T__EXCHANGE_ERROR}

		for i := 0; i < len(permitted); i++ {
			if s == permitted[i] {
				return nil
			}
		}

		return _errors.New("is invalid")
	}
}

func nStatusValidation(s int) validation.RuleFunc {
	return func(value interface{}) error {
		permitted := []int{static.NTF__S__NEW, static.NTF__S__IN_PROCESS, static.NTF__S__COMPLETED}

		for i := 0; i < len(permitted); i++ {
			if s == permitted[i] {
				return nil
			}
		}

		return _errors.New("is invalid")
	}
}
