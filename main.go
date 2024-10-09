package main

import (
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/gin-gonic/gin"
	"github.com/go-gorp/gorp/v3"
	"github.com/google/uuid"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/rtgraph"
	"github.com/minor-industries/rtgraph/schema"
	"github.com/pkg/errors"
	"html/template"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
	"weight-tracker/assets"
	"weight-tracker/db"
)

const (
	dbHost   = "127.0.0.1"
	location = "America/Los_Angeles"

	day = 24 * time.Hour

	defaultStartDate = "2011-01-01"
)

var opts struct {
	Listen     string `long:"listen" default:"0.0.0.0:8000" env:"LISTEN"`
	DisableTLS bool   `long:"disable-tls" env:"DISABLE_TLS"`
	TLSCert    string `long:"tls-cert" env:"TLS_CERT" default:"/var/lib/tls/certs/server.pem"`
	TLSKey     string `long:"tls-key" env:"TLS_KEY" default:"/var/lib/tls/private/server-key.pem"`
}

type StorageBackend struct {
	db *gorp.DbMap
}

func (s *StorageBackend) LoadDataWindow(
	seriesName string,
	start time.Time,
) (schema.Series, error) {
	var values []schema.Value

	switch seriesName {
	case "weight":
		rows, err := getDataAfter(s.db, start)
		if err != nil {
			return schema.Series{}, errors.Wrap(err, "get data after")
		}

		values = make([]schema.Value, len(rows))

		for idx, row := range rows {
			values[idx] = schema.Value{
				Timestamp: row.T,
				Value:     row.Weight,
			}
		}
	default:
		return schema.Series{}, errors.New("unknown series")
	}

	return schema.Series{
		SeriesName: seriesName,
		Values:     values,
	}, nil
}

func (s *StorageBackend) CreateSeries(seriesNames []string) error {
	return nil
}

func (s *StorageBackend) InsertValue(seriesName string, timestamp time.Time, value float64) error {
	//TODO implement me
	panic("implement me")
}

func run() error {
	_, err := flags.Parse(&opts)
	if err != nil {
		return errors.Wrap(err, "parse flags")
	}

	dbmap, err := db.Get(dbHost)
	if err != nil {
		return errors.Wrap(err, "get db")
	}

	backend := &StorageBackend{
		db: dbmap,
	}

	errCh := make(chan error)
	graph, err := rtgraph.New(
		backend,
		errCh,
		rtgraph.Opts{},
		[]string{"weight"},
	)
	if err != nil {
		return errors.Wrap(err, "new rtgraph")
	}

	graph.StaticFiles(assets.FS,
		"purecss/base-min.css", "text/css",
		"purecss/pure-min.css", "text/css",
	)

	graph.GetEngine().GET("/ios-icon.png", func(c *gin.Context) {
		c.FileFromFS("/ios-icon.png", http.FS(assets.FS))
	})

	r := graph.GetEngine()

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

	r.GET("/index.html", func(c *gin.Context) {
		after, err := time.Parse("2006-01-02", c.DefaultQuery("after", defaultStartDate))
		if err != nil {
			c.AbortWithError(400, errors.Wrap(err, "parse time"))
			return
		}

		data, err := getDataAfter(dbmap, after)
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

	r.GET("/data.csv", func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/plain")
		c.Status(200)

		after, err := time.Parse("2006-01-02", c.DefaultQuery("after", defaultStartDate))
		if err != nil {
			c.AbortWithError(400, errors.Wrap(err, "parse time"))
			return
		}

		data, err := getDataAfter(dbmap, after)
		if err != nil {
			_ = c.Error(err)
			return
		}

		lines := []string{"date,weight"}

		for i, d := range data {
			if i > 0 {
				d0 := data[i-1]
				dt := d.T.Sub(d0.T)
				if dt > 7*day {
					lines = append(lines, fmt.Sprintf("%s,NaN",
						d0.T.Format("2006/01/02 15:04:05"),
					))
				}
			}

			lines = append(lines, fmt.Sprintf("%s,%f",
				d.T.Format("2006/01/02 15:04:05"),
				d.Weight,
			))
		}

		content := []byte(strings.Join(lines, "\n"))

		_, _ = c.Writer.Write(content)
	})

	if opts.DisableTLS {
		if err := r.Run(opts.Listen); err != nil {
			return errors.Wrap(err, "run")
		}
	} else {
		if err := r.RunTLS(opts.Listen, opts.TLSCert, opts.TLSKey); err != nil {
			return errors.Wrap(err, "run")
		}
	}

	return nil
}

func files(r *gin.Engine, files ...string) {
	for i := 0; i < len(files); i += 2 {
		name := files[i]
		ct := files[i+1]
		r.GET("/"+name, func(c *gin.Context) {
			header := c.Writer.Header()
			header["Content-Type"] = []string{ct}
			content, err := fs.ReadFile(assets.FS, name)
			if err != nil {
				c.Status(404)
				return
			}
			_, _ = c.Writer.Write(content)
		})
	}
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
