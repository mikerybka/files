package files

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Server struct {
	BaseDir string
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGet(w, r)
	case http.MethodPut:
		s.handlePut(w, r)
	case http.MethodDelete:
		s.handleDelete(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGet serves files from baseDir or returns a JSON directory listing.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	// Convert URL path to a local path under baseDir
	path := filepath.Join(s.BaseDir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))

	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "404 not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// If it's a directory, return a JSON list of dir entries
	if info.IsDir() {
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
func (s *Server) handlePut(w http.ResponseWriter, r *http.Request) {
	// Convert URL path to a local file path under baseDir
	path := filepath.Join(s.BaseDir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))

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

	// Copy the request body into the file
	if _, err := io.Copy(file, r.Body); err != nil {
		http.Error(w, "Error writing file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// handleDelete removes the file or directory at the given path.
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	// Convert URL path to a local path under baseDir
	path := filepath.Join(s.BaseDir, filepath.FromSlash(strings.TrimPrefix(r.URL.Path, "/")))

	// Check if path exists
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Remove the file or directory (recursively, if a directory)
	if err := os.RemoveAll(path); err != nil {
		http.Error(w, "Could not delete file or directory", http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusOK)
}
