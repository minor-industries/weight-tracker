package main

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/go-gorp/gorp/v3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"html/template"
	"net/http"
	"sort"
	"strconv"
	"time"
	"weight-tracker/assets"
	"weight-tracker/db"
	"weight-tracker/graphs"
)

const (
	dbHost   = "127.0.0.1"
	location = "America/Los_Angeles"
)

func run() error {
	r := gin.Default()

	dbmap, err := db.Get(dbHost)
	if err != nil {
		return errors.Wrap(err, "get db")
	}

	funcs := map[string]any{
		"Localtime": func(w db.Weight) string {
			loc, err := time.LoadLocation(w.Location)
			if err != nil {
				return "~location error~"
			}
			return w.T.In(loc).Format("2006-01-02 15:04:05")
		},
		"FmtWeight": func(w float64) string {
			return fmt.Sprintf("%.02f", w)
		},
		"Delta": func(w []db.Weight, i0 int) string {
			i1 := i0 + 1
			if i1 >= len(w) {
				return "-"
			}

			delta := w[i0].Weight - w[i1].Weight

			return fmt.Sprintf("%+0.02f", delta)
		},
		"DateOfWeek": func(w db.Weight) string {
			loc, err := time.LoadLocation(w.Location)
			if err != nil {
				return "~location error~"
			}
			return w.T.In(loc).Format("Mon")
		},

		"DaysMissing": func(w []db.Weight, i0 int) string {
			i1 := i0 + 1
			if i1 >= len(w) {
				return ""
			}

			loc0, err := time.LoadLocation(w[i0].Location)
			if err != nil {
				return "~location error~"
			}

			loc1, err := time.LoadLocation(w[i1].Location)
			if err != nil {
				return "~location error~"
			}

			t0 := w[i0].T.In(loc0)
			t1 := w[i1].T.In(loc1)

			day0 := time.Date(t0.Year(), t0.Month(), t0.Day(), 0, 0, 0, 0, time.UTC)
			day1 := time.Date(t1.Year(), t1.Month(), t1.Day(), 0, 0, 0, 0, time.UTC)

			delta := day0.Sub(day1)
			days := int(delta.Hours() / 24)

			switch days {
			case 0, 1:
				return ""
			case 2:
				return "1 day missing"
			default:
				return fmt.Sprintf("%d days missing", days)
			}
		},
	}

	templ := template.Must(template.New("").Funcs(funcs).ParseFS(assets.FS, "*.html"))
	r.SetHTMLTemplate(templ)

	r.GET("/", func(c *gin.Context) {
		months := getMonthsParam(c, 3)

		data, err := getData(dbmap, months)
		if err != nil {
			_ = c.Error(err)
			return
		}

		sort.Slice(data, func(i, j int) bool {
			return data[i].T.After(data[j].T)
		})

		c.HTML(http.StatusOK, "index.html", map[string]any{
			"date":   time.Now().String(),
			"action": "form-handler",
			"id":     uuid.New().String(),
			"data":   data,
		})
	})

	r.GET("/weight.svg", func(c *gin.Context) {
		months := getMonthsParam(c, 3)

		data, err := getData(dbmap, months)
		if err != nil {
			_ = c.Error(err)
			return
		}

		svg, err := graphs.Graph(data)
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

	r.GET("/commit-and-push.html", func(c *gin.Context) {
		commit, err := db.CommitAndPush(dbmap)

		args := map[string]any{
			"err": err,
		}

		if commit.Valid {
			args["commit"] = commit.String
		}

		c.HTML(200, "commit-and-push.html", args)
	})

	if err := r.Run("0.0.0.0:8000"); err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}

func getMonthsParam(c *gin.Context, dflt int) int {
	months := dflt
	if m, ok := c.GetQuery("months"); ok {
		mInt, err := strconv.ParseInt(m, 10, 64)
		switch err {
		case nil:
			months = int(mInt)
		}
	}
	return months
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
