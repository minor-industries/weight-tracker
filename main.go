package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
	"time"
)

//go:embed *.html
var f embed.FS

func run() error {
	r := gin.Default()

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
		weight, _ := c.GetPostForm("weight")
		unit, _ := c.GetPostForm("unit")

		c.HTML(http.StatusInternalServerError, "form-handler.html", map[string]any{
			"weight": weight,
			"unit":   unit,
		})
	})

	err := r.Run("127.0.0.1:8000")
	if err != nil {

	}

	return nil
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
