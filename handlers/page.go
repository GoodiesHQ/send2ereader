package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/goodieshq/send2ereader/helpers"
	"github.com/goodieshq/send2ereader/templates"
)

// newSession creates a new session and redirects to the transfer page
func (h *Handler) newSession(w http.ResponseWriter, r *http.Request) {
	sess, err := h.store.CreateCode(h.codeSize)
	if err != nil {
		http.Error(w, "unable to create session code", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/transfer/"+sess.Code, http.StatusFound)
}

// Index detects the device's UA and routes accordingly.
// Known E-readers get a fresh session code, everything else sees the upload form.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	// If the request comes from a known ereader, redirect to a new session page
	if helpers.IsRequestFromEreader(r) {
		h.newSession(w, r)
	} else {
		// Otherwise, show the computer file upload page.
		render(w, r, templates.ComputerPage(h.codeSize))
	}
}

// Receive creates a new session and redirects to the transfer page
func (h *Handler) Receive(w http.ResponseWriter, r *http.Request) {
	// Manually make a new session and redirect to the transfer page
	h.newSession(w, r)
}

// EreaderPage is the ereader landing page
func (h *Handler) EreaderPage(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	// Check if the code is valid
	sess, ok := h.store.Get(code)
	if !ok {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// If the file is ready, show the download link
	downloadURL := ""
	if _, ready := sess.GetFile(); ready {
		downloadURL = "/download/" + code
	}

	render(w, r, templates.EreaderPage(sess.Code, r.Host, downloadURL, helpers.IsRequestFromEreader(r)))
}
