package main

import (
	"embed"
	"github.com/gin-gonic/gin"
	"html/template"
	"net/http"
)

//go:embed *.html
var f embed.FS

func run() error {
	r := gin.Default()

	templ := template.Must(template.New("").ParseFS(f, "*.html"))
	r.SetHTMLTemplate(templ)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
		//c.String(200, "no")
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
