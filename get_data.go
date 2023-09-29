package main

import (
	"github.com/go-gorp/gorp/v3"
	"github.com/pkg/errors"
	"weight-tracker/db"
)

var query = `
SELECT
	*
FROM
	weight
WHERE
	t >= DATE_SUB(CURDATE(), INTERVAL :months MONTH)
ORDER BY
	t ASC
`

func getData(dbmap *gorp.DbMap, months int) ([]db.Weight, error) {
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
