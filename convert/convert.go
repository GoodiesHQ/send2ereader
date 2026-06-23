package convert

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/pgaskin/kepubify/v4/kepub"
)

type Format string

// KepubifyName takes a filename and returns the appropriate KEPUB filename.
func KepubifyName(name string) (string, bool) {
	// normalize the name to lowercase for extension checking, but preserve the original case for the returned name
	nameLower := strings.ToLower(name)

	// If the name already ends with .kepub.epub, return it unchanged
	if strings.HasSuffix(nameLower, ".kepub.epub") {
		return name, false
	}

	// If the name ends with .epub, replace it with .kepub.epub
	if strings.HasSuffix(nameLower, ".epub") {
		return replaceExt(name, "kepub.epub"), true
	}

	// not an epub, don't convert
	return name, false
}

// toKepub converts an EPUB []byte to a KEPUB []byte using kepubify.
// The returned name replaces the .epub extension with .kepub.epub.
func Kepubify(data []byte, inputName string) ([]byte, string, error) {
	name, changed := KepubifyName(inputName)
	if !changed {
		// If the name didn't change, it's not an EPUB or already a KEPUB, so return the original data and name.
		return data, name, nil
	}

	// Parse the EPUB data as a zip archive
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, "", fmt.Errorf("parsing epub: %w", err)
	}

	// Convert to KEPUB using the kepubify library
	var out bytes.Buffer
	if err := kepub.NewConverter().Convert(context.Background(), &out, zr); err != nil {
		return nil, "", fmt.Errorf("kepubify: %w", err)
	}

	// Return the converted data and the new filename
	return out.Bytes(), name, nil
}

// replaceExt replaces or appends the file extension of name with ext
func replaceExt(name, ext string) string {
	if i := strings.LastIndex(name, "."); i >= 0 {
		return name[:i+1] + ext
	}
	return name + "." + ext
}
