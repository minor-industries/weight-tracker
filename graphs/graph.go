package graphs

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
	Timezone string    `db:"timezone,size:255"`
	Weight   float64   `db:"weight"`
	Unit     string    `db:"unit,size:32"`
}

func Graph() error {
	const host = "127.0.0.1"
	url := fmt.Sprintf("root:@tcp(%s)/weight?parseTime=true", host)

	db, err := sql.Open("mysql", url)
	if err != nil {
		return errors.Wrap(err, "open db")
	}

	if err := db.Ping(); err != nil {
		return errors.Wrap(err, "ping")
	}

	// construct a gorp DbMap
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{}}
	fmt.Println(dbmap)

	var data []Weight
	if _, err := dbmap.Select(&data, "select * from weight order by t"); err != nil {
		return errors.Wrap(err, "select")
	}

	for _, w := range data {
		fmt.Println(w)
	}

	return nil
}
