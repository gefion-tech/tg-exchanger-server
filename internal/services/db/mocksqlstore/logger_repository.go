package mocksqlstore

import (
	"database/sql"
	"time"

	"github.com/gefion-tech/tg-exchanger-server/internal/core"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
)

type LoggerRepository struct {
	logs map[int]*models.LogRecord
}

func (r *LoggerRepository) Create(lr *models.LogRecord) error {
	lr.ID = len(r.logs) + 1
	lr.CreatedAt = time.Now().UTC().Format(core.DateStandart)

	r.logs[lr.ID] = lr
	return nil
}

func (r *LoggerRepository) Delete(lr *models.LogRecord) error {
	if r.logs[lr.ID] != nil {
		r.rewrite(lr.ID, lr)
		defer delete(r.logs, r.logs[lr.ID].ID)
		return nil
	}

	return sql.ErrNoRows
}

func (r *LoggerRepository) Count(querys interface{}) (int, error) {
	return len(r.logs), nil
}

func (r *LoggerRepository) Selection(querys interface{}) ([]*models.LogRecord, error) {
	arr := []*models.LogRecord{}
	for _, lr := range r.logs {
		arr = append(arr, lr)
	}

	return arr, nil
}

func (r *LoggerRepository) DeleteSelection(querys interface{}) ([]*models.LogRecord, error) {
	arr := []*models.LogRecord{}
	for _, lr := range r.logs {
		arr = append(arr, lr)
	}

	return arr, nil
}

func (r *LoggerRepository) rewrite(id int, to *models.LogRecord) {
	to.ID = r.logs[id].ID
	to.Username = r.logs[id].Username
	to.Info = r.logs[id].Info
	to.Service = r.logs[id].Service
	to.Module = r.logs[id].Module
	to.CreatedAt = r.logs[id].CreatedAt
}
