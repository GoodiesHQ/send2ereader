package helpers

import (
	"net/http"
	"strings"
)

// Brand represents a known e-reader brand
type Brand string

const (
	BrandUnknown Brand = "unknown"
	BrandKobo    Brand = "kobo"
	BrandAmazon  Brand = "amazon"
)

// Detection holds the result of a user-agent detection
type Detection struct {
	Brand     Brand
	UserAgent string
}

// Detect analyzes the request's user-agent and returns a Detection struct
func Detect(r *http.Request) Detection {
	if r == nil {
		return Detection{Brand: BrandUnknown}
	}

	// Get and normalize the user-agent string for easier detection
	ua := r.UserAgent()
	uaLower := strings.ToLower(ua)

	switch {
	case strings.Contains(uaLower, "kobo"):
		return Detection{
			Brand:     BrandKobo,
			UserAgent: ua,
		}
	case strings.Contains(uaLower, "kindle") ||
		strings.Contains(uaLower, "silk") ||
		strings.Contains(uaLower, "kfapwi") ||
		strings.Contains(uaLower, "kftt") ||
		strings.Contains(uaLower, "kfot") ||
		strings.Contains(uaLower, "kfjwa") ||
		strings.Contains(uaLower, "kfsowi"):
		return Detection{
			Brand:     BrandAmazon,
			UserAgent: ua,
		}
	default:
		return Detection{
			Brand:     BrandUnknown,
			UserAgent: ua,
		}
	}
}

// IsRequestFromEreader returns true if the request's user-agent matches a known e-reader brand
func IsRequestFromEreader(r *http.Request) bool {
	d := Detect(r)
	return d.Brand != BrandUnknown
}

// IsRequestFromKobo returns true if the request's user-agent matches a Kobo e-reader
func IsRequestFromKobo(r *http.Request) bool {
	return Detect(r).Brand == BrandKobo
}

// IsRequestFromAmazon returns true if the request's user-agent matches an Amazon Kindle e-reader
func IsRequestFromAmazon(r *http.Request) bool {
	return Detect(r).Brand == BrandAmazon
}
