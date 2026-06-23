package handlers

import (
	"log"

	"github.com/goodieshq/send2ereader/store"
)

// File size limit
const maxUploadBytes = 100 << 20 // 100 MB

// Handler is the main HTTP handler struct, maintains a session store
type Handler struct {
	store    *store.Store
	codeSize int
}

// New creates a new Handler with the given session store and code size
func New(store *store.Store, codeSize int) *Handler {
	log.Printf("Using code size: %d", codeSize)
	return &Handler{store, codeSize}
}
