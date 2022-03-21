package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.LoadHTMLGlob("index.html")
	sData := "paste it below"
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData})
	})

	r.POST("/", func(c *gin.Context) {
		sData = c.PostForm("sdata")
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData})
	})
	fmt.Println("server start at http://127.0.0.1:7777")

	r.Run(":7777")
}
