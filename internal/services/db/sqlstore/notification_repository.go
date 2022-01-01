package sqlstore

import (
	"database/sql"
	"time"

	"github.com/gefion-tech/tg-exchanger-server/internal/models"
	"github.com/gefion-tech/tg-exchanger-server/internal/tools"
)

type NotificationRepository struct {
	store *sql.DB
}

/*
	Создать записи в таблице `notifications`

	# TESTED
*/
func (r *NotificationRepository) Create(n *models.Notification) error {
	if err := r.store.QueryRow(
		`
		INSERT INTO notifications(type, chat_id, username, code, user_card, img_path, ex_from, ex_to)
		SELECT $1, $2, $3, $4, $5, $6, $7, $8
		RETURNING id, type, status, chat_id, username, code, user_card, img_path, ex_from, ex_to, created_at, updated_at
		`,
		n.Type,
		n.User.ChatID,
		n.User.Username,
		n.MetaData.CardVerification.Code,
		n.MetaData.CardVerification.UserCard,
		n.MetaData.CardVerification.ImgPath,
		n.MetaData.ExActionCancel.ExFrom,
		n.MetaData.ExActionCancel.ExTo,
	).Scan(
		&n.ID,
		&n.Type,
		&n.Status,
		&n.User.ChatID,
		&n.User.Username,
		&n.MetaData.CardVerification.Code,
		&n.MetaData.CardVerification.UserCard,
		&n.MetaData.CardVerification.ImgPath,
		&n.MetaData.ExActionCancel.ExFrom,
		&n.MetaData.ExActionCancel.ExTo,
		&n.CreatedAt,
		&n.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

/*
	Обновить поле status в таблице `notifications`

	# TESTED
*/
func (r *NotificationRepository) Update(n *models.Notification) error {
	if err := r.store.QueryRow(
		`
		UPDATE notifications
		SET status=$1, updated_at=$2
		WHERE id=$3
		RETURNING id, type, status, chat_id, username, code, user_card, img_path, ex_from, ex_to, created_at, updated_at
		`,
		n.Status,
		time.Now().UTC().Format("2006-01-02T15:04:05.00000000"),
		n.ID,
	).Scan(
		&n.ID,
		&n.Type,
		&n.Status,
		&n.User.ChatID,
		&n.User.Username,
		&n.MetaData.CardVerification.Code,
		&n.MetaData.CardVerification.UserCard,
		&n.MetaData.CardVerification.ImgPath,
		&n.MetaData.ExActionCancel.ExFrom,
		&n.MetaData.ExActionCancel.ExTo,
		&n.CreatedAt,
		&n.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}

/*
	Получить кол-во записей `notifications`

	# TESTED
*/
func (r *NotificationRepository) Count() (int, error) {
	var c int
	if err := r.store.QueryRow(
		`
		SELECT count(*)
		FROM notifications		
		`,
	).Scan(
		&c,
	); err != nil {
		return 0, err
	}

	return c, nil
}

/*
	Получить выборку из таблицы `notifications`

	# TESTED
*/
func (r *NotificationRepository) Selection(page, limit int) ([]*models.Notification, error) {
	nArr := []*models.Notification{}

	rows, err := r.store.Query(
		`
		SELECT id, type, status, chat_id, username, code, user_card, img_path, ex_from, ex_to, created_at, updated_at
		FROM notifications
		ORDER BY id DESC
		OFFSET $1
		LIMIT $2
		`,
		tools.OffsetThreshold(page, limit),
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		n := models.Notification{}
		if err := rows.Scan(
			&n.ID,
			&n.Type,
			&n.Status,
			&n.User.ChatID,
			&n.User.Username,
			&n.MetaData.CardVerification.Code,
			&n.MetaData.CardVerification.UserCard,
			&n.MetaData.CardVerification.ImgPath,
			&n.MetaData.ExActionCancel.ExFrom,
			&n.MetaData.ExActionCancel.ExTo,
			&n.CreatedAt,
			&n.UpdatedAt,
		); err != nil {
			continue
		}

		nArr = append(nArr, &n)
	}

	return nArr, nil
}

/*
	Получить запись из таблицы `notifications`

	# TESTED
*/
func (r *NotificationRepository) Get(n *models.Notification) error {
	if err := r.store.QueryRow(
		`
		SELECT id, type, status, chat_id, username, code, user_card, img_path, ex_from, ex_to, created_at, updated_at
		FROM notifications
		WHERE id=$1
		`,
		n.ID,
	).Scan(
		&n.ID,
		&n.Type,
		&n.Status,
		&n.User.ChatID,
		&n.User.Username,
		&n.MetaData.CardVerification.Code,
		&n.MetaData.CardVerification.UserCard,
		&n.MetaData.CardVerification.ImgPath,
		&n.MetaData.ExActionCancel.ExFrom,
		&n.MetaData.ExActionCancel.ExTo,
		&n.CreatedAt,
		&n.UpdatedAt,
	); err != nil {
		return err
	}
	return nil
}

/*
	Удалить запись из таблицы `notifications`

	# TESTED
*/
func (r *NotificationRepository) Delete(n *models.Notification) error {
	if err := r.store.QueryRow(
		`
		DELETE FROM notifications
		WHERE id=$1
		RETURNING id, type, status, chat_id, username, code, user_card, img_path, ex_from, ex_to, created_at, updated_at
		`,
		n.ID,
	).Scan(
		&n.ID,
		&n.Type,
		&n.Status,
		&n.User.ChatID,
		&n.User.Username,
		&n.MetaData.CardVerification.Code,
		&n.MetaData.CardVerification.UserCard,
		&n.MetaData.CardVerification.ImgPath,
		&n.MetaData.ExActionCancel.ExFrom,
		&n.MetaData.ExActionCancel.ExTo,
		&n.CreatedAt,
		&n.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}
