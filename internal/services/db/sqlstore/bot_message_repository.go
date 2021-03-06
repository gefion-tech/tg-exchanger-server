package sqlstore

import (
	"database/sql"
	"time"

	"github.com/gefion-tech/tg-exchanger-server/internal/core"
	AppMath "github.com/gefion-tech/tg-exchanger-server/internal/core/math"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
)

type BotMessagesRepository struct {
	store *sql.DB
}

/*
	Создать сообщение в таблице `bot_messages`

	# TESTED
*/
func (r *BotMessagesRepository) Create(m *models.BotMessage) error {
	if err := r.store.QueryRow(
		`
		INSERT INTO bot_messages (connector, message_text, created_by)
		SELECT $1, $2, $3
		WHERE NOT EXISTS (SELECT connector FROM bot_messages WHERE connector=$4)
		RETURNING id, connector, message_text, created_by, created_at, updated_at
		`,
		m.Connector,
		m.MessageText,
		m.CreatedBy,
		m.Connector,
	).Scan(
		&m.ID,
		&m.Connector,
		&m.MessageText,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

func (r *BotMessagesRepository) Count(querys interface{}) (int, error) {
	var c int
	if err := r.store.QueryRow(
		`
		SELECT count(*)
		FROM bot_messages		
		`,
	).Scan(
		&c,
	); err != nil {
		return 0, err
	}

	return c, nil
}

/*
	Получить конкретное сообщение из таблицы `bot_messages`

	# TESTED
*/
func (r *BotMessagesRepository) Get(m *models.BotMessage) error {
	if err := r.store.QueryRow(
		`
		SELECT id, connector, message_text, created_by, created_at, updated_at
		FROM bot_messages WHERE connector=$1
		`,
		m.Connector,
	).Scan(
		&m.ID,
		&m.Connector,
		&m.MessageText,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

/*
	Получить выборку из таблицы `bot_messages`
*/
func (r *BotMessagesRepository) Selection(querys interface{}) ([]*models.BotMessage, error) {
	bmArr := []*models.BotMessage{}
	q := querys.(*models.BotMessageSelection)

	rows, err := r.store.Query(
		`
		SELECT id, connector, message_text, created_by, created_at, updated_at
		FROM bot_messages
		ORDER BY id DESC
		OFFSET $1
		LIMIT $2
		`,
		AppMath.OffsetThreshold(q.Page, q.Limit),
		q.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		m := &models.BotMessage{}
		if err := rows.Scan(
			&m.ID,
			&m.Connector,
			&m.MessageText,
			&m.CreatedBy,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			continue
		}

		bmArr = append(bmArr, m)
	}

	return bmArr, nil
}

/*
	Обновить конкретное сообщение в таблице `bot_messages`

	# TESTED
*/
func (r *BotMessagesRepository) Update(m *models.BotMessage) error {
	if err := r.store.QueryRow(
		`
		UPDATE bot_messages
		SET message_text=$1, updated_at=$2
		WHERE id=$3
		RETURNING id, connector, message_text, created_by, created_at, updated_at
		`,
		m.MessageText,
		time.Now().UTC().Format(core.DateStandart),
		m.ID,
	).Scan(
		&m.ID,
		&m.Connector,
		&m.MessageText,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}

/*
	Удалить конкретное сообщение в таблице `bot_messages`
*/
func (r *BotMessagesRepository) Delete(m *models.BotMessage) error {
	if err := r.store.QueryRow(
		`
		DELETE FROM bot_messages
		WHERE id=$1
		RETURNING id, connector, message_text, created_by, created_at, updated_at
		`,
		m.ID,
	).Scan(
		&m.ID,
		&m.Connector,
		&m.MessageText,
		&m.CreatedBy,
		&m.CreatedAt,
		&m.UpdatedAt,
	); err != nil {
		return err
	}

	return nil
}
