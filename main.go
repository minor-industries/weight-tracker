package main

import (
	"embed"
	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/go-gorp/gorp/v3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"html/template"
	"net/http"
	"strconv"
	"time"
	"weight-tracker/db"
	"weight-tracker/graphs"
)

const (
	dbHost   = "127.0.0.1"
	location = "America/Los_Angeles"
)

//go:embed *.html
var f embed.FS

func run() error {
	r := gin.Default()

	dbmap, err := db.Get(dbHost)
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

	r.GET("/weight.svg", func(c *gin.Context) {
		months := 3
		if m, ok := c.GetQuery("months"); ok {
			mInt, err := strconv.ParseInt(m, 10, 64)
			switch err {
			case nil:
				months = int(mInt)
			}
		}

		svg, err := graphs.Graph(dbmap, months)
		if err != nil {
			panic(err)
		}
		c.Data(200, "image/svg+xml", svg)
	})

	r.POST("/form-handler", func(c *gin.Context) {
		res, err := writeWeightToDB(
			dbmap,
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

	if err := r.Run("0.0.0.0:8000"); err != nil {
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
	dbmap *gorp.DbMap,
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

	id, err := uuid.Parse(idParam)
	if err != nil {
		return nil, errors.Wrap(err, "parse id")
	}

	idBytes, err := id.MarshalBinary()
	if err != nil {
		return nil, errors.Wrap(err, "marshal id")
	}

	w := &db.Weight{
		Id:       idBytes,
		T:        now.UTC(),
		Location: location,
		Weight:   weight,
		Unit:     unitParam,
	}

	err = dbmap.Insert(w)
	if err != nil {
		return nil, errors.Wrap(err, "insert")
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
