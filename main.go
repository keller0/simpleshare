package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

const TmpFileDir = "tmpFile"

func init() {
	tDir := filepath.Join(".", TmpFileDir)
	err := os.MkdirAll(tDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
}
func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.LoadHTMLGlob("index.html")
	sData := "paste it below"
	var imgList []string

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData, "imaList": imgList})
	})

	r.POST("/", func(c *gin.Context) {
		sData = c.PostForm("sdata")
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData, "imaList": imgList})

	})

	// Set memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 50 << 20

	r.POST("/upload", func(c *gin.Context) {
		// Single file
		file, err := c.FormFile("file")
		if err != nil {
			fmt.Println(err)
			c.String(http.StatusBadRequest, err.Error())
		}

		saveName := strconv.FormatInt(time.Now().UnixNano(), 10) + path.Ext(file.Filename)

		err = c.SaveUploadedFile(file, TmpFileDir+"/"+saveName)
		if err != nil {
			fmt.Println(err)
			c.String(http.StatusBadRequest, err.Error())
		}
		fileUrl := "tFile/" + saveName
		imgList = append([]string{fileUrl}, imgList...)

		c.String(http.StatusOK, fileUrl)
	})
	r.Static("/tFile", "./"+TmpFileDir)

	fmt.Println("server start at http://127.0.0.1:7777")

	r.Run(":7777")
}
