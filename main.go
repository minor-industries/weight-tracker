package main

import (
	"embed"
	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"weight/db"
)

//go:embed *.html
var f embed.FS

func run() error {
	r := gin.Default()

	conn, err := db.Get("git01.vpn")
	if err != nil {
		return errors.Wrap(err, "get db")
	}

	templ := template.Must(template.New("").ParseFS(f, "*.html"))
	r.SetHTMLTemplate(templ)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", map[string]any{
			"date":   time.Now().String(),
			"action": "form-handler",
			"id":     uuid.New().String(),
		})
	})

	r.POST("/form-handler", func(c *gin.Context) {
		res, err := writeWeightToDB(
			conn,
			c.PostForm("weight"),
			c.PostForm("unit"),
			c.PostForm("id"),
		)
		if err != nil {
			c.HTML(400, "error.html", map[string]any{
				"message": err.Error(),
			})
			return
		}

		c.HTML(http.StatusOK, "form-handler.html", res)
	})

	if err := r.Run("127.0.0.1:8000"); err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}

func Exprs(first string, rest ...string) []any {
	result := []any{
		squirrel.Expr(first),
	}
	for _, i := range rest {
		result = append(result, squirrel.Expr(i))
	}
	return result
}

func writeWeightToDB(
	conn *sqlx.DB,
	weightParam string,
	unitParam string,
	idParam string,
) (map[string]any, error) {
	switch unitParam {
	case "kg", "lbs":
		//pass
	default:
		return nil, errors.New("invalid unit")
	}

	weight, err := strconv.ParseFloat(weightParam, 64)
	if err != nil {
		return nil, errors.Wrap(err, "parse weight")
	}

	now := time.Now()
	zone, _ := now.Zone()

	sql, _, err := squirrel.Insert("weight").Columns(
		"id",
		"t",
		"timezone",
		"weight",
		"unit",
	).Values(Exprs(
		"uuid_to_bin(:id)",
		":t",
		":timezone",
		":weight",
		":unit",
	)...).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "to sql")
	}

	query, args, err := sqlx.Named(sql, &db.Weight{
		Id:       []byte(idParam),
		T:        now.UTC(),
		Timezone: zone,
		Weight:   weight,
		Unit:     unitParam,
	})
	if err != nil {
		return nil, errors.Wrap(err, "named")
	}

	query = conn.Rebind(query)

	if _, err := conn.Exec(query, args...); err != nil {
		return nil, errors.Wrap(err, "exec")
	}

	return map[string]any{
		"weight": weight,
		"unit":   unitParam,
		"t":      now,
		"id":     idParam,
	}, nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
