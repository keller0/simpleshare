package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

const TmpFileDir = "tmpFile"

func init() {
	tDir := filepath.Join(".", TmpFileDir)
	err := os.MkdirAll(tDir, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

type tmpFile struct {
	T  string `json:"t"`
	N  string `json:"n"`
	Ns string `json:"ns"`
}

func main() {
	r := gin.Default()
	r.SetTrustedProxies(nil)

	r.LoadHTMLGlob("index.html")
	sData := ""
	var imgList []string
	var fileList []tmpFile

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData, "imaList": imgList, "fileList": fileList})
	})

	r.POST("/", func(c *gin.Context) {
		sData = c.PostForm("sData")
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData, "imaList": imgList, "fileList": fileList})
	})

	r.POST("/clearFile", func(c *gin.Context) {
		fileList = nil
		imgList = nil
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData, "imaList": imgList, "fileList": fileList})
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

		v, e := c.GetPostForm("isImg")
		if !e {
			v = "0"
		}
		if v == "1" {
			imgList = append([]string{fileUrl}, imgList...)
		} else {
			fileList = append(fileList, tmpFile{T: "0", N: fileUrl, Ns: file.Filename})
		}

		c.String(http.StatusOK, fileUrl)
	})

	r.Static("/tFile", "./"+TmpFileDir)
	r.Static("/static", "./static")

	fmt.Println("server start at http://127.0.0.1:7777")

	r.Run(":7777")
}
