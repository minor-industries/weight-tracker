package db

import (
	"database/sql"
	"github.com/go-gorp/gorp/v3"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

func CommitAndPush(dbmap *gorp.DbMap) (sql.NullString, error) {
	var commit sql.NullString
	_, err := dbmap.SelectInt(`call dolt_add('.')`)
	if err != nil {
		return commit, errors.Wrap(err, "add")
	}

	commit, err = dbmap.SelectNullStr(`call dolt_commit('-m', 'data updates')`)
	if err != nil {
		var mysqlErr *mysql.MySQLError
		ok := errors.As(err, &mysqlErr)

		if ok && mysqlErr.Number == 1105 && mysqlErr.Message == "nothing to commit" {
			// ignore this error, it's ok
		} else {
			return commit, errors.Wrap(err, "commit")
		}
	}

	_, err = dbmap.SelectInt(`call dolt_push('origin', 'main')`)
	if err != nil {
		return commit, errors.Wrap(err, "push")
	}

	return commit, nil
}
