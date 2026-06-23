package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/goodieshq/send2ereader/convert"
	"github.com/goodieshq/send2ereader/helpers"
	"github.com/goodieshq/send2ereader/store"
	"github.com/goodieshq/send2ereader/templates"
)

// Poll is the HTMX partials endpoint for non-e-reader browsers
func (h *Handler) Poll(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// If the code is malformed, show an invalid message.
	if len(code) != h.codeSize {
		render(w, r, templates.PollInvalid())
		return
	}

	sess, ok := h.store.Get(code)
	// If the session doesn't exist or the code is invalid, show an invalid message.
	if !ok {
		render(w, r, templates.PollInvalid())
		return
	}

	// If the session exists but no file is ready, show the waiting message.
	f, ready := sess.GetFile()
	if !ready {
		render(w, r, templates.PollWaiting(code))
		return
	}

	// File is ready, show the download link.
	name := f.Name
	if f.Kepubify && helpers.IsRequestFromKobo(r) {
		name, _ = convert.KepubifyName(name)
	}

	render(w, r, templates.PollReady(code, name))
}

// Upload handles the computer-side multipart form submission.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	// Limit the size of the request body to prevent abuse
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

	// Parse the multipart form with a size limit to prevent abuse.
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		render(w, r, templates.UploadError("File is too large (100 MB maximum)."))
		return
	}

	// Normalize and validate the session code from the form
	code := strings.ToUpper(strings.TrimSpace(r.FormValue("code")))
	if len(code) != h.codeSize {
		w.WriteHeader(http.StatusBadRequest)
		render(w, r, templates.UploadError(fmt.Sprintf("Invalid %d-character code.", h.codeSize)))
		return
	}

	// Look up the session. If it doesn't exist or is expired, show an error.
	sess, ok := h.store.Get(code)
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		render(w, r, templates.UploadError("Code is invalid or expired. Check the code on your e-reader and try again."))
		return
	}

	// Read the uploaded file from the form
	f, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		render(w, r, templates.UploadError("No file selected."))
		return
	}
	defer f.Close()

	// Read the file data into memory with a size limit
	data, err := io.ReadAll(io.LimitReader(f, maxUploadBytes))
	if err != nil {
		http.Error(w, "reading file", http.StatusInternalServerError)
		return
	}

	// Convert the file to the requested format (if any)
	filename := path.Base(header.Filename)

	// Store the file in the session
	if err := sess.SetFile(&store.File{
		Name:     filename,
		Data:     data,
		Kepubify: r.FormValue("kepubify") == "on",
	}); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		render(w, r, templates.UploadError("The file has already been uploaded for this session."))
		return
	}

	// Show the success message with a download link to the e-reader page
	render(w, r, templates.UploadSuccess(code))
}

// Status returns a JSON payload that the e-reader page's ES5 polling script reads.
// {"ready":false}                                — file not yet uploaded
// {"ready":true,"filename":"book.epub"}          — file is waiting
// 404                                            — session expired/unknown
func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Check the session for the code. If it doesn't exist, return 404.
	sess, ok := h.store.Get(code)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"session not found"}`))
		return
	}

	// Return the file status as JSON. If a file is ready, include the filename for display on the e-reader page.
	w.Header().Set("Content-Type", "application/json")
	if f, ready := sess.GetFile(); ready {
		name := f.Name
		if f.Kepubify && helpers.IsRequestFromKobo(r) {
			name, _ = convert.KepubifyName(name)
		}
		fmt.Fprintf(w, `{"ready":true,"filename":%q}`, name)
	} else {
		_, _ = w.Write([]byte(`{"ready":false}`))
	}
}

// Download serves the stored file
func (h *Handler) Download(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Check the session for the code. If it doesn't exist, return 404.
	sess, ok := h.store.Get(code)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// If the session exists but no file is ready, return 404.
	f, ready := sess.GetFile()
	if !ready {
		http.Error(w, "file not ready", http.StatusNotFound)
		return
	}

	// If the device is a Kobo and the file should be kepubified, convert it before serving.
	if helpers.IsRequestFromKobo(r) {
		// Check if the file should be converted to KEPUB format
		if f.Kepubify {
			converted, outName, err := convert.Kepubify(f.Data, f.Name)
			if err != nil {
				log.Printf("failed to convert to KEPUB, serving EPUB: %v", err)
			} else {
				f.Data = converted
				f.Name = outName
			}
		}
	}

	reader := bytes.NewReader(f.Data)
	cd := mime.FormatMediaType("attachment", map[string]string{"filename": f.Name})
	w.Header().Set("Content-Type", mimeFor(f.Name))
	w.Header().Set("Content-Disposition", cd)
	http.ServeContent(w, r, f.Name, sess.CreatedAt, reader)
}
