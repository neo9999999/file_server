package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Directory to store uploaded files
const uploadDir = "./uploads/"

// Log file path
const logFile = "./server.log"

// Ensure upload directory exists
func init() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
}

// Log request details
func logRequest(r *http.Request) {
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening log file:", err)
		return
	}
	defer f.Close()

	logEntry := fmt.Sprintf("%s - %s %s %s\n", time.Now().Format(time.RFC3339), r.RemoteAddr, r.Method, r.URL.Path)
	if _, err := f.WriteString(logEntry); err != nil {
		log.Println("Error writing to log file:", err)
	}
}

// Upload file handler
func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to upload file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	out, err := os.Create(filepath.Join(uploadDir, filename))
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	logRequest(r)
	fmt.Fprintf(w, "File %s uploaded successfully!", filename)
}

// Download file handler
func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")
	filePath := filepath.Join(uploadDir, fileName)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	http.ServeFile(w, r, filePath)
	logRequest(r)
}

// Simple home page handler
func homePage(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html lang="en">
		<head><title>File Server</title></head>
		<body>
			<h1>Upload a File</h1>
			<form enctype="multipart/form-data" action="/upload" method="post">
				<input type="file" name="file" />
				<input type="submit" value="Upload" />
			</form>
			<h1>Download a File</h1>
			<form action="/download" method="get">
				<input type="text" name="file" placeholder="Filename" />
				<input type="submit" value="Download" />
			</form>
		</body>
		</html>`)
}

func main() {
	// Set up handlers
	http.HandleFunc("/", homePage)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/download", downloadFile)

	// Start server
	fmt.Println("Server starting at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
