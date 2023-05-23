package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type tmpFile struct {
	T  string `json:"t"`
	N  string `json:"n"`
	Ns string `json:"ns"`
}

var (
	sData    string
	imgList  []string
	fileList []tmpFile
)

func runServer() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.SetTrustedProxies(nil)

	setRouter(r)

	// Set memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 50 << 20

	ads := address + ":" + port
	fmt.Println("server started at http://" + ads)

	loadOldFiles()

	return r.Run(ads)

}

func setRouter(r *gin.Engine) {

	r.LoadHTMLGlob("index.html")

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
	r.POST("/upload", func(c *gin.Context) {
		// Single file
		file, err := c.FormFile("file")
		if err != nil {
			fmt.Println(err)
			c.String(http.StatusBadRequest, err.Error())
		}

		// saveName := strconv.FormatInt(time.Now().UnixNano(), 10) + path.Ext(file.Filename)
		saveName := file.Filename

		err = c.SaveUploadedFile(file, tmpFileDir+"/"+saveName)
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

	r.Static("/tFile", "./"+tmpFileDir)
	r.Static("/static", "./static")

}

// load files in tmpFileDir
func loadOldFiles() {
	files, err := ioutil.ReadDir(tmpFileDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		// fmt.Println(f.Name())

		fileUrl := "tFile/" + f.Name()
		if isImgSimple(f.Name()) {
			imgList = append([]string{fileUrl}, imgList...)

		} else {
			fileList = append(fileList, tmpFile{T: "0", N: fileUrl, Ns: f.Name()})

		}
	}

}

func isImgSimple(name string) bool {
	if filepath.Ext(name) == ".jpg" || filepath.Ext(name) == ".png" || filepath.Ext(name) == ".jpeg" {
		return true
	}
	return false
}
