package main

import (
	"github.com/go-gorp/gorp/v3"
	"github.com/pkg/errors"
	"time"
	"weight-tracker/db"
)

const query = `
SELECT
	*
FROM
	weight
WHERE
	t >= DATE_SUB(CURDATE(), INTERVAL :months MONTH)
ORDER BY
	t ASC
`

func getData_(dbmap *gorp.DbMap, months int) ([]db.Weight, error) {
	var data []db.Weight
	if _, err := dbmap.Select(
		&data,
		query,
		map[string]any{"months": months},
	); err != nil {
		return nil, errors.Wrap(err, "select")
	}

	return data, nil
}

const queryAfter = `
SELECT
	*
FROM
	weight
WHERE
	t >= :t
ORDER BY
	t ASC
`

func getDataAfter(dbmap *gorp.DbMap, t time.Time) ([]db.Weight, error) {
	var data []db.Weight
	if _, err := dbmap.Select(
		&data,
		queryAfter,
		map[string]any{"t": t},
	); err != nil {
		return nil, errors.Wrap(err, "select")
	}

	return data, nil
}
