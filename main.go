package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/goodieshq/send2ereader/handlers"
	"github.com/goodieshq/send2ereader/store"
)

// Settings represents the global application settings
type Settings struct {
	CodeSize   int
	Expiration time.Duration
	Port       uint16
}

// Default settings
const defaultPort = uint16(8080)
const defaultCodeSize = 4
const defaultTimeout = 10 * time.Minute

func main() {
	// Get settings from environment variables or use defaults
	settings := getSettings()

	// Create the main handler with the session store and code size
	store := store.New(settings.Expiration)

	// Periodically purge expired sessions
	go func() {
		// Run cleanup every minute
		t := time.NewTicker(1 * time.Minute)
		defer t.Stop()
		for range t.C {
			store.Cleanup()
		}
	}()

	// Create a new session store (in-memory by default, may be replaced with a persistent store in the future)
	h := handlers.New(store, settings.CodeSize)

	// Set up the router with middleware
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	// Serve static files from the embedded static/ directory at /static/*
	r.Handle("/static/*", http.StripPrefix("/static", staticHandler()))

	// Define application routes
	r.Get("/", h.Index)                      // device detection and routing
	r.Get("/receive", h.Receive)             // manual override: any device can act as receiver
	r.Get("/transfer/{code}", h.EreaderPage) // ereader landing page (waiting for file or showing download link)
	r.Get("/status/{code}", h.Status)        // polled by the e-reader page ES5 script
	r.Get("/poll/{code}", h.Poll)            // HTMX partials for desktop browsers
	r.Post("/upload", h.Upload)              // computer-side file upload endpoint
	r.Get("/download/{code}", h.Download)    // file download endpoint for ereaders

	// Start the server
	addr := ":" + strconv.Itoa(int(settings.Port))
	log.Printf("send2ereader listening on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

//go:embed static
var staticEmbed embed.FS

// staticHandler serves files from the embedded static/ directory at /static/*.
func staticHandler() http.Handler {
	sub, err := fs.Sub(staticEmbed, "static")
	if err != nil {
		panic(err)
	}
	return http.FileServer(http.FS(sub))
}

// getSettings reads the configuration from environment variables
func getSettings() Settings {
	codeSize := defaultCodeSize
	if envCodeSize := os.Getenv("CODE_SIZE"); envCodeSize != "" {
		if cs, err := strconv.Atoi(envCodeSize); err == nil && cs > 0 {
			codeSize = cs
		} else {
			log.Printf("Invalid CODE_SIZE value: %q. Using default: %d", envCodeSize, defaultCodeSize)
		}
	}

	timeout := defaultTimeout
	if envTimeout := os.Getenv("EXPIRATION_MINUTES"); envTimeout != "" {
		if tm, err := strconv.Atoi(envTimeout); err == nil && tm > 0 {
			timeout = time.Duration(tm) * time.Minute
		} else {
			log.Printf("Invalid EXPIRATION_MINUTES value: %q. Using default: %v", envTimeout, defaultTimeout)
		}
	}

	port := defaultPort
	if envPort := os.Getenv("PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil && p > 0 && p < 65536 {
			port = uint16(p)
		} else {
			log.Printf("Invalid PORT value: %q. Using default :%d", envPort, defaultPort)
		}
	}

	return Settings{
		CodeSize:   codeSize,
		Expiration: timeout,
		Port:       port,
	}
}
