package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const baseDir = "./data"

// DirEntry represents a single file or directory in JSON.
type DirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func main() {
	http.HandleFunc("/", handleRequest)

	log.Println("Serving files and accepting PUT requests at http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGet(w, r)
	case http.MethodPut:
		handlePut(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet serves files from baseDir or returns a JSON directory listing.
func handleGet(w http.ResponseWriter, r *http.Request) {
	// Convert URL path to a local path under baseDir
	path := filepath.Join(baseDir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "404 not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if info.IsDir() {
		// Return JSON list of dir entries
		files, err := os.ReadDir(path)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		entries := []DirEntry{}
		for _, f := range files {
			entryType := "file"
			if f.IsDir() {
				entryType = "dir"
			}

			entries = append(entries, DirEntry{
				Name: f.Name(),
				Type: entryType,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(entries); err != nil {
			http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		}
		return
	}

	// If it's a file, just serve it
	http.ServeFile(w, r, path)
}

// handlePut writes the request body to a file under baseDir.
func handlePut(w http.ResponseWriter, r *http.Request) {
	// Convert URL path to a local file path under baseDir
	path := filepath.Join(baseDir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))

	// Create intermediate directories if necessary
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		http.Error(w, "Could not create directories", http.StatusInternalServerError)
		return
	}

	// Create or overwrite the file
	file, err := os.Create(path)
	if err != nil {
		http.Error(w, "Could not create file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Copy request body into the file
	if _, err := io.Copy(file, r.Body); err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
