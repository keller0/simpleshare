package main

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

//go:embed web
var webDir embed.FS

// tmpFile struct
type tmpFile struct {
	T  string `json:"t"`  // 0: normal file
	N  string `json:"n"`  // file url
	Ns string `json:"ns"` // file name
}

var (
	sData    string
	imgList  []string  // img url list
	fileList []tmpFile // file list
)

func runServer() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	err := r.SetTrustedProxies(nil)
	if err != nil {
		panic(err)
	}

	setRouter(r)

	loadOldFiles()

	// Set memory limit for multipart forms
	r.MaxMultipartMemory = 20 << 20 // 20 MiB

	ads := address + ":" + port
	fmt.Println("server started at http://" + ads)

	return r.Run(ads)

}

func setRouter(r *gin.Engine) {

	tempHtml := template.Must(template.New("").ParseFS(webDir, "web/*.html"))
	r.SetHTMLTemplate(tempHtml)
	r.LoadHTMLGlob("web/*.html")
	// r.Static("/static", "./static")
	r.StaticFS("/static", http.FS(webDir))
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData, "imgList": imgList, "fileList": fileList})
	})

	r.POST("/", func(c *gin.Context) {
		sData = c.PostForm("sData")
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData, "imgList": imgList, "fileList": fileList})
	})

	r.POST("/clearAll", func(c *gin.Context) {
		sData = ""
		clearTmpFile()
		fileList = nil
		imgList = nil
		c.HTML(http.StatusOK, "index.html", gin.H{"data": sData, "imgList": imgList, "fileList": fileList})
	})

	r.POST("/deleteFile", func(c *gin.Context) {
		sData = ""
		fileName := c.PostForm("fileName")
		if fileName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fileName is required"})
			return
		}

		// 删除文件
		filePath := filepath.Join(tmpFileDir, fileName)
		err := os.Remove(filePath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file: " + err.Error()})
			return
		}

		// 从列表中移除
		fileUrl := "tFile/" + fileName
		if isImgSimple(fileName) {
			// 从图片列表中移除
			for i, img := range imgList {
				if img == fileUrl {
					imgList = append(imgList[:i], imgList[i+1:]...)
					break
				}
			}
		} else {
			// 从文件列表中移除
			for i, file := range fileList {
				if file.N == fileUrl {
					fileList = append(fileList[:i], fileList[i+1:]...)
					break
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{"success": true, "message": "File deleted successfully"})
	})

	r.POST("/upload", func(c *gin.Context) {
		// Single file
		file, err := c.FormFile("file")
		if err != nil {
			fmt.Println(err)
			c.String(http.StatusBadRequest, err.Error())
		}

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

}

// load files in tmpFileDir
func loadOldFiles() {
	files, err := os.ReadDir(tmpFileDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fileUrl := "tFile/" + f.Name()
		if isImgSimple(f.Name()) {
			imgList = append([]string{fileUrl}, imgList...)
		} else {
			fileList = append(fileList, tmpFile{T: "0", N: fileUrl, Ns: f.Name()})
		}
	}
}

func clearTmpFile() {
	err := os.RemoveAll(tmpFileDir)
	if err != nil {
		log.Fatal(err)
	}
	err = os.MkdirAll(tmpFileDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func isImgSimple(name string) bool {
	if filepath.Ext(name) == ".jpg" || filepath.Ext(name) == ".png" || filepath.Ext(name) == ".jpeg" || filepath.Ext(name) == ".webp" {
		return true
	}
	return false
}
