package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func Get() (*sqlx.DB, error) {
	conn, err := sqlx.Connect("mysql", "root:@/weight?parseTime=true")
	if err != nil {
		return nil, errors.Wrap(err, "connect")
	}

	if _, err := conn.Exec(`
		create table
			if not exists 
		weight(
			id binary(16),
		    t datetime,
			weight double,
			unit varchar(32),
		    PRIMARY KEY(id)
		);`,
	); err != nil {
		return nil, err
	}

	return conn, err
}
