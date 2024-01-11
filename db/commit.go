package db

import (
	"github.com/go-gorp/gorp/v3"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

func CommitAndPush(dbmap *gorp.DbMap) error {
	_, err := dbmap.SelectInt(`call dolt_add('.')`)
	if err != nil {
		return errors.Wrap(err, "add")
	}

	_, err = dbmap.SelectStr(`call dolt_commit('-m', 'data updates')`)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		ok := errors.As(err, &mysqlErr)

		if ok && mysqlErr.Number == 1105 {
			// ignore this case: 1105 => nothing to commit
		} else {
			return errors.Wrap(err, "commit")
		}
	}

	_, err = dbmap.SelectInt(`call dolt_push('origin', 'main')`)
	if err != nil {
		return errors.Wrap(err, "add")
	}

	return nil
}
