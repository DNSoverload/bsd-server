package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"archive/zip"
	"io"
)

const downloadDir = "./" // Folder with files to serve

func main() {
	// Route: list files
	http.HandleFunc("/", listFilesHandler)

	// Route: download individual file
	http.HandleFunc("/download/", downloadFileHandler)
	http.HandleFunc("/download-folder", zipFolder)
	fmt.Println("Serving on http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(downloadDir)
	if err != nil {
		http.Error(w, "Failed to read dir", 500)
		return
	}

	fmt.Fprintln(w, "<h2>Download Files</h2><ul>")
	for _, file := range files {
		if !file.IsDir() {
			name := file.Name()
			link := "/download/" + name
			fmt.Fprintf(w, `<li><a href="%s">%s</a></li>`, link, name)
		}
	}
	fmt.Fprintln(w, "</ul>")
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/download/")
	fullpath := filepath.Join(downloadDir, filename)

	

	http.ServeFile(w, r, fullpath)
}

func zipFolder(w http.ResponseWriter, r *http.Request) {
	folder := r.URL.Query().Get("folder")
	basePath := "."

	fullPath := filepath.Join(basePath, filepath.Clean(folder))

	// Security: ensure within base dir
	if !strings.HasPrefix(fullPath, filepath.Clean(basePath)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filepath.Base(folder)+".zip\"")

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Create zip path
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}
		relPath = filepath.ToSlash(relPath)

		if info.IsDir() {
			if relPath == "." {
				return nil
			}
			_, err = zipWriter.Create(relPath + "/")
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		writer, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		return err
	})

	if err != nil {
		http.Error(w, "Failed to zip folder: "+err.Error(), 500)
	}
}
