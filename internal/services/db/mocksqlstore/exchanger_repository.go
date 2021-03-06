package mocksqlstore

import (
	"database/sql"
	"time"

	"github.com/gefion-tech/tg-exchanger-server/internal/core"
	AppMath "github.com/gefion-tech/tg-exchanger-server/internal/core/math"
	"github.com/gefion-tech/tg-exchanger-server/internal/models"
)

type ExchangerRepository struct {
	exchangers map[int]*models.Exchanger
}

func (r *ExchangerRepository) Create(e *models.Exchanger) error {
	e.ID = len(r.exchangers) + 1
	e.CreatedAt = time.Now().UTC().Format(core.DateStandart)
	e.UpdatedAt = time.Now().UTC().Format(core.DateStandart)

	r.exchangers[e.ID] = e
	return nil
}

func (r *ExchangerRepository) Update(e *models.Exchanger) error {
	if r.exchangers[e.ID] != nil {
		r.exchangers[e.ID].Name = e.Name
		r.exchangers[e.ID].UrlToParse = e.UrlToParse
		r.exchangers[e.ID].UpdatedAt = time.Now().UTC().Format(core.DateStandart)
		return nil

	}

	for _, ex := range r.exchangers {
		if ex.ID == e.ID {
			r.exchangers[e.ID].Name = e.Name
			r.exchangers[e.ID].UrlToParse = e.UrlToParse
			r.exchangers[e.ID].UpdatedAt = time.Now().UTC().Format(core.DateStandart)

			r.rewrite(ex.ID, e)
			return nil
		}
	}

	return sql.ErrNoRows
}

func (r *ExchangerRepository) GetByName(e *models.Exchanger) error {
	for _, ex := range r.exchangers {
		if ex.Name == e.Name {
			return nil
		}
	}

	return sql.ErrNoRows
}

func (r *ExchangerRepository) Delete(e *models.Exchanger) error {
	if r.exchangers[e.ID] != nil {
		defer delete(r.exchangers, r.exchangers[e.ID].ID)
		return nil
	}

	return sql.ErrNoRows
}

func (r *ExchangerRepository) Count(querys interface{}) (int, error) {
	return len(r.exchangers), nil
}

func (r *ExchangerRepository) Selection(querys interface{}) ([]*models.Exchanger, error) {
	arr := []*models.Exchanger{}
	q := querys.(*models.ExchangerSelection)

	for i, v := range r.exchangers {
		if i > AppMath.OffsetThreshold(q.Page, q.Limit) && i <= AppMath.OffsetThreshold(q.Page, q.Limit)+q.Limit {
			arr = append(arr, v)
		}
		i++
	}

	return arr, nil
}

func (r *ExchangerRepository) GetSlice(limit int) ([]*models.Exchanger, error) {
	eArr := []*models.Exchanger{}

	for i := 0; i < limit; i++ {
		eArr = append(eArr, r.exchangers[i])
	}

	return eArr, nil
}

func (r *ExchangerRepository) rewrite(id int, to *models.Exchanger) {
	to.ID = r.exchangers[id].ID
	to.Name = r.exchangers[id].Name
	to.UrlToParse = r.exchangers[id].UrlToParse
	to.CreatedBy = r.exchangers[id].CreatedBy
	to.CreatedAt = r.exchangers[id].CreatedAt
	to.UpdatedAt = r.exchangers[id].UpdatedAt
}
