package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
)
func main() {
	r := gin.Default()

	r.LoadHTMLGlob("index.html")
	sData := "paste it below"
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"data" : sData})
	})

	r.POST("/", func(c *gin.Context) {
		sData = c.PostForm("sdata")
		c.HTML(http.StatusOK, "index.html", gin.H{"data" : sData})
	})

	r.Run(":7777")
}