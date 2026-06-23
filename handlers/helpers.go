package handlers

import (
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/a-h/templ"
)

// mimeFor returns the appropriate MIME type based on the file extension.
func mimeFor(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".epub":
		return "application/epub+zip"
	case ".mobi", ".azw", ".azw3":
		return "application/x-mobipocket-ebook"
	case ".pdf":
		return "application/pdf"
	case ".txt":
		return "text/plain; charset=utf-8"
	case ".html", ".htm":
		return "text/html; charset=utf-8"
	case ".cbz":
		return "application/x-cbz"
	case ".cbr":
		return "application/x-cbr"
	default:
		return "application/octet-stream"
	}
}

// render is a helper for rendering templates and logging any errors
func render(w http.ResponseWriter, r *http.Request, c templ.Component) {
	if err := c.Render(r.Context(), w); err != nil {
		log.Printf("render error: %v", err)
	}
}
