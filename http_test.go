package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetState(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	tmpFileDir = dir

	stateMu.Lock()
	sData = ""
	imgList = nil
	fileList = nil
	stateMu.Unlock()

	return dir
}

func newTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	mux, err := setRouter()
	if err != nil {
		t.Fatalf("setRouter() error = %v", err)
	}
	return mux
}

func TestSanitizeFileName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{name: "plain name", input: "notes.txt", want: "notes.txt"},
		{name: "trim spaces", input: "  photo.png  ", want: "photo.png"},
		{name: "strip unix path", input: "../../etc/passwd", want: "passwd"},
		{name: "strip windows path", input: `..\..\secret.txt`, want: "secret.txt"},
		{name: "empty", input: "   ", wantErr: true},
		{name: "dot", input: ".", wantErr: true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := sanitizeFileName(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("sanitizeFileName(%q) expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("sanitizeFileName(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("sanitizeFileName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsImgSimple(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want bool
	}{
		{name: "a.jpg", want: true},
		{name: "a.jpeg", want: true},
		{name: "a.png", want: true},
		{name: "a.webp", want: true},
		{name: "a.txt", want: false},
		{name: "a.PNG", want: false},
		{name: "noext", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isImgSimple(tt.name); got != tt.want {
				t.Fatalf("isImgSimple(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestIndexPage(t *testing.T) {
	resetState(t)
	mux := newTestMux(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("GET / status = %d, want %d", rec.Code, http.StatusOK)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "simple share") {
		t.Fatalf("GET / body missing page title, got: %s", body)
	}
}

func TestSubmitText(t *testing.T) {
	resetState(t)
	mux := newTestMux(t)

	form := url.Values{}
	form.Set("sData", "hello shared text")
	req := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /submit status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("success = %v, want true", resp["success"])
	}
	if resp["data"] != "hello shared text" {
		t.Fatalf("data = %v, want hello shared text", resp["data"])
	}

	stateMu.RLock()
	defer stateMu.RUnlock()
	if sData != "hello shared text" {
		t.Fatalf("sData = %q, want hello shared text", sData)
	}
}

func TestUploadAndServeFile(t *testing.T) {
	dir := resetState(t)
	mux := newTestMux(t)

	fileURL := uploadFile(t, mux, "hello.txt", []byte("file-content"), "0")
	if fileURL != "tFile/hello.txt" {
		t.Fatalf("upload url = %q, want tFile/hello.txt", fileURL)
	}

	saved := filepath.Join(dir, "hello.txt")
	content, err := os.ReadFile(saved)
	if err != nil {
		t.Fatalf("read uploaded file: %v", err)
	}
	if string(content) != "file-content" {
		t.Fatalf("uploaded content = %q, want file-content", content)
	}

	stateMu.RLock()
	if len(fileList) != 1 || fileList[0].Ns != "hello.txt" {
		stateMu.RUnlock()
		t.Fatalf("fileList = %+v, want one hello.txt entry", fileList)
	}
	stateMu.RUnlock()

	req := httptest.NewRequest(http.MethodGet, "/"+fileURL, nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /%s status = %d, want %d", fileURL, rec.Code, http.StatusOK)
	}
	if rec.Body.String() != "file-content" {
		t.Fatalf("served content = %q, want file-content", rec.Body.String())
	}
}

func TestUploadImage(t *testing.T) {
	resetState(t)
	mux := newTestMux(t)

	fileURL := uploadFile(t, mux, "pic.png", []byte("png-bytes"), "1")
	if fileURL != "tFile/pic.png" {
		t.Fatalf("upload url = %q, want tFile/pic.png", fileURL)
	}

	stateMu.RLock()
	defer stateMu.RUnlock()
	if len(imgList) != 1 || imgList[0] != fileURL {
		t.Fatalf("imgList = %+v, want [%s]", imgList, fileURL)
	}
	if len(fileList) != 0 {
		t.Fatalf("fileList = %+v, want empty", fileList)
	}
}

func TestUploadRejectsInvalidFileName(t *testing.T) {
	resetState(t)
	mux := newTestMux(t)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", ".")
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := part.Write([]byte("x")); err != nil {
		t.Fatalf("write part: %v", err)
	}
	_ = writer.WriteField("isImg", "0")
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /upload status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestDeleteFile(t *testing.T) {
	dir := resetState(t)
	mux := newTestMux(t)
	_ = uploadFile(t, mux, "delete-me.txt", []byte("bye"), "0")

	form := url.Values{}
	form.Set("fileName", "delete-me.txt")
	req := httptest.NewRequest(http.MethodPost, "/deleteFile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /deleteFile status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("success = %v, want true", resp["success"])
	}

	if _, err := os.Stat(filepath.Join(dir, "delete-me.txt")); !os.IsNotExist(err) {
		t.Fatalf("expected file removed, stat err = %v", err)
	}

	stateMu.RLock()
	defer stateMu.RUnlock()
	if len(fileList) != 0 {
		t.Fatalf("fileList = %+v, want empty after delete", fileList)
	}
}

func TestDeleteFileMissing(t *testing.T) {
	resetState(t)
	mux := newTestMux(t)

	form := url.Values{}
	form.Set("fileName", "missing.txt")
	req := httptest.NewRequest(http.MethodPost, "/deleteFile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("POST /deleteFile status = %d, want %d; body=%s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestDeleteFileRejectsEmptyName(t *testing.T) {
	resetState(t)
	mux := newTestMux(t)

	form := url.Values{}
	form.Set("fileName", "   ")
	req := httptest.NewRequest(http.MethodPost, "/deleteFile", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST /deleteFile status = %d, want %d; body=%s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestClearAll(t *testing.T) {
	dir := resetState(t)
	mux := newTestMux(t)

	_ = uploadFile(t, mux, "a.txt", []byte("a"), "0")
	_ = uploadFile(t, mux, "b.png", []byte("b"), "1")

	form := url.Values{}
	form.Set("sData", "to-clear")
	submitReq := httptest.NewRequest(http.MethodPost, "/submit", strings.NewReader(form.Encode()))
	submitReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	submitRec := httptest.NewRecorder()
	mux.ServeHTTP(submitRec, submitReq)
	if submitRec.Code != http.StatusOK {
		t.Fatalf("setup submit status = %d", submitRec.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/clearAll", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /clearAll status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("tmp dir entries = %d, want 0", len(entries))
	}

	stateMu.RLock()
	defer stateMu.RUnlock()
	if sData != "" || len(fileList) != 0 || len(imgList) != 0 {
		t.Fatalf("state not cleared: sData=%q fileList=%+v imgList=%+v", sData, fileList, imgList)
	}
}

func TestLoadOldFiles(t *testing.T) {
	dir := resetState(t)
	if err := os.WriteFile(filepath.Join(dir, "old.txt"), []byte("old"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "old.png"), []byte("img"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	loadOldFiles()

	stateMu.RLock()
	defer stateMu.RUnlock()
	if len(fileList) != 1 || fileList[0].Ns != "old.txt" {
		t.Fatalf("fileList = %+v, want one old.txt", fileList)
	}
	if len(imgList) != 1 || imgList[0] != "tFile/old.png" {
		t.Fatalf("imgList = %+v, want [tFile/old.png]", imgList)
	}
}

func TestMethodNotAllowed(t *testing.T) {
	resetState(t)
	mux := newTestMux(t)

	endpoints := []string{"/submit", "/clearAll", "/deleteFile", "/upload"}
	for _, endpoint := range endpoints {
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("GET %s status = %d, want %d", endpoint, rec.Code, http.StatusMethodNotAllowed)
		}
	}
}

func uploadFile(t *testing.T, mux http.Handler, filename string, content []byte, isImg string) string {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		t.Fatalf("copy content: %v", err)
	}
	if err := writer.WriteField("isImg", isImg); err != nil {
		t.Fatalf("WriteField: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("POST /upload status = %d, want %d; body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	return rec.Body.String()
}
