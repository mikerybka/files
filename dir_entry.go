package files

// DirEntry represents a single file or directory in JSON.
type DirEntry struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
