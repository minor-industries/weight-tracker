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
			"date": time.Now().String(),
			"id":   "abc123",
		})
	})

	r.POST("/weight-form", func(c *gin.Context) {
		w, _ := c.GetPostForm("w")

		c.HTML(http.StatusInternalServerError, "form-submit.html", map[string]any{
			"w": w,
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
