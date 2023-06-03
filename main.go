package main

import (
	"embed"
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

	conn, err := db.Get()
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
		if err := writeWeightToDB(c, conn); err != nil {
			c.HTML(400, "error.html", map[string]any{
				"message": err.Error(),
			})
			return
		}
	})

	if err := r.Run("127.0.0.1:8000"); err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}

func writeWeightToDB(c *gin.Context, conn *sqlx.DB) error {
	weightParam := c.PostForm("weight")
	unitParam := c.PostForm("unit")
	idParam := c.PostForm("id")

	switch unitParam {
	case "kg", "lbs":
		//pass
	default:
		return errors.New("invalid unit")
	}

	weight, err := strconv.ParseFloat(weightParam, 64)
	if err != nil {
		return errors.Wrap(err, "parse weight")
	}

	// TODO: validate form values
	now := time.Now()
	zone, _ := now.Zone()

	_, err = conn.NamedExec(
		`insert into weight(
		   	id,
		   	t,
		   	timezone,
		   	weight, 
			unit
		) values (
			uuid_to_bin(:id), 
			:t, 
			:timezone, 
			:weight, 
			:unit
		)`,
		&db.Weight{
			Id:       idParam,
			T:        now.UTC(),
			Timezone: zone,
			Weight:   weight,
			Unit:     unitParam,
		})
	if err != nil {
		return errors.Wrap(err, "insert")
	}

	c.HTML(http.StatusInternalServerError, "form-handler.html", map[string]any{
		"weight": weight,
		"unit":   unitParam,
		"t":      now,
		"id":     idParam,
	})

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
