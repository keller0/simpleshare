package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

type tmpFile struct {
	T  string `json:"t"`
	N  string `json:"n"`
	Ns string `json:"ns"`
}

var (
	address    string
	port       string
	tmpFileDir string
)

func runServer() error {
	gin.SetMode(gin.ReleaseMode)
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

	ads := address + ":" + port
	fmt.Println("server started at http://" + ads)
	return r.Run(ads)

}

var rootCmd = &cobra.Command{
	Use:   "simpleshare",
	Short: "Share texts and files in local network",
	Long: `
  A Simple http service for share texts and files in local network
built in Go.
  source: https://github.com/keller0/simpleshare`,
	CompletionOptions: cobra.CompletionOptions{HiddenDefaultCmd: true},
	Version:           "0.0.2",
	Example:           "./simpleshare  -a 127.0.0.1 -p 7777 -f tmpFile",
	Run: func(cmd *cobra.Command, args []string) {
		tDir := filepath.Join(".", tmpFileDir)
		err := os.MkdirAll(tDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
		fmt.Println("tmp file folder:", tDir)
		err = runServer()
		if err != nil {
			fmt.Println(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "7777", "listen port")
	rootCmd.PersistentFlags().StringVarP(&address, "address", "a", "127.0.0.1", "listen address")
	rootCmd.PersistentFlags().StringVarP(&tmpFileDir, "folder", "f", "tmpFile", "tmp file folder")

}

func main() {
	Execute()
}
