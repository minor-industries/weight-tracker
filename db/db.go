package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"time"
)

type Weight struct {
	Id       []byte    `db:"id"`
	T        time.Time `db:"t"`
	Timezone string    `db:"timezone"`
	Weight   float64   `db:"weight"`
	Unit     string    `db:"unit"`
}

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
		    timezone varchar(255),
			weight double,
			unit varchar(32),
		    PRIMARY KEY(id)
		);`,
	); err != nil {
		return nil, err
	}

	return conn, err
}
