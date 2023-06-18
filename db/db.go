package db

import (
	"database/sql"
	"fmt"
	"github.com/go-gorp/gorp/v3"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"time"
)

type Weight struct {
	Id       []byte    `db:"id"`
	T        time.Time `db:"t"`
	Location string    `db:"location,size:255"`
	Weight   float64   `db:"weight"`
	Unit     string    `db:"unit,size:32"`
}

func Get(dbHost string) (*gorp.DbMap, error) {
	url := fmt.Sprintf("root:@tcp(%s)/weight?parseTime=true", dbHost)

	db, err := sql.Open("mysql", url)
	if err != nil {
		return nil, errors.Wrap(err, "open db")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "ping")
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
	dbmap.AddTableWithName(Weight{}, "weight")

	return dbmap, nil
}
