package main

import (
	"embed"
	"github.com/gin-gonic/gin"
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
			"id":     "abc123",
			"action": "form-handler",
		})
	})

	r.POST("/form-handler", func(c *gin.Context) {
		weightParam, _ := c.GetPostForm("weight")
		unitParam, _ := c.GetPostForm("unit")

		weight, _ := strconv.ParseFloat(weightParam, 64)

		// TODO: validate form values
		_, err := conn.NamedExec(`insert into weight(id, t, weight, unit) values(:id, :t, :weight, :unit)`, &db.Weight{
			Id:     []byte{},
			T:      time.Now(),
			Weight: weight,
			Unit:   unitParam,
		})

		c.HTML(http.StatusInternalServerError, "form-handler.html", map[string]any{
			"weight":  weight,
			"unit":    unitParam,
			"message": err.Error(),
		})
	})

	if err := r.Run("127.0.0.1:8000"); err != nil {
		return errors.Wrap(err, "run")
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
