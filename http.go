package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
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
	stateMu  sync.RWMutex
)

func runServer() error {
	mux, err := setRouter()
	if err != nil {
		return err
	}

	loadOldFiles()

	ads := address + ":" + port
	fmt.Println("server started at http://" + ads)

	srv := &http.Server{
		Addr:    ads,
		Handler: mux,
	}

	return srv.ListenAndServe()

}

func setRouter() (*http.ServeMux, error) {
	tempHTML, err := template.New("").ParseFS(webDir, "web/*.html")
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(webDir))))
	mux.Handle("/tFile/", http.StripPrefix("/tFile/", http.FileServer(http.Dir(tmpFileDir))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		renderIndex(w, tempHTML)
	})

	mux.HandleFunc("/submit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}

		stateMu.Lock()
		sData = r.PostFormValue("sData")
		resp := map[string]any{
			"success":  true,
			"data":     sData,
			"imgList":  append([]string(nil), imgList...),
			"fileList": append([]tmpFile(nil), fileList...),
		}
		stateMu.Unlock()

		writeJSON(w, http.StatusOK, resp)
	})

	mux.HandleFunc("/clearAll", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		stateMu.Lock()
		sData = ""
		clearTmpFile()
		fileList = nil
		imgList = nil
		stateMu.Unlock()
		renderIndex(w, tempHTML)
	})

	mux.HandleFunc("/deleteFile", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}

		fileName, err := sanitizeFileName(r.PostFormValue("fileName"))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}

		filePath := filepath.Join(tmpFileDir, fileName)
		if err := os.Remove(filePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				writeJSON(w, http.StatusNotFound, map[string]any{"error": "file not found"})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "failed to delete file: " + err.Error()})
			return
		}

		fileURL := "tFile/" + fileName
		stateMu.Lock()
		if isImgSimple(fileName) {
			for i, img := range imgList {
				if img == fileURL {
					imgList = append(imgList[:i], imgList[i+1:]...)
					break
				}
			}
		} else {
			for i, file := range fileList {
				if file.N == fileURL {
					fileList = append(fileList[:i], fileList[i+1:]...)
					break
				}
			}
		}
		stateMu.Unlock()

		writeJSON(w, http.StatusOK, map[string]any{"success": true, "message": "File deleted successfully"})
	})

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(20 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		saveName, err := sanitizeFileName(fileHeader.Filename)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dst, err := os.OpenFile(filepath.Join(tmpFileDir, saveName), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fileURL := "tFile/" + saveName
		if r.PostFormValue("isImg") == "1" {
			stateMu.Lock()
			imgList = append([]string{fileURL}, imgList...)
			stateMu.Unlock()
		} else {
			stateMu.Lock()
			fileList = append(fileList, tmpFile{T: "0", N: fileURL, Ns: saveName})
			stateMu.Unlock()
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(fileURL))
	})

	return mux, nil
}

// load files in tmpFileDir
func loadOldFiles() {
	files, err := os.ReadDir(tmpFileDir)
	if err != nil {
		log.Fatal(err)
	}

	stateMu.Lock()
	defer stateMu.Unlock()
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

func renderIndex(w http.ResponseWriter, tpl *template.Template) {
	stateMu.RLock()
	data := map[string]any{
		"data":     sData,
		"imgList":  append([]string(nil), imgList...),
		"fileList": append([]tmpFile(nil), fileList...),
	}
	stateMu.RUnlock()

	if err := tpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func sanitizeFileName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", errors.New("fileName is required")
	}

	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	baseName := path.Base(normalized)
	if baseName == "." || baseName == "/" || baseName == "" {
		return "", errors.New("invalid file name")
	}
	if strings.Contains(baseName, "/") || strings.Contains(baseName, string(os.PathSeparator)) {
		return "", errors.New("invalid file name")
	}

	return baseName, nil
}
