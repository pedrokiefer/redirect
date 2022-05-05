package redirect

import "net/http"

// Engine of all redirection.
type Engine interface {
	http.Handler
	Reload() error // reload configuration from storage
}

// Stats consumer.
type StatWriter interface {
	Touch(url string) // Touch resource and increment counter (hot operation, should be fast)
}

// Stats reader.
type StatReader interface {
	Visits(url string) int64 // Get number of visits for specific service/url
}

// Stats reader and writer.
type Stats interface {
	StatWriter
	StatReader
}

// Single rule for redirection.
type Rule struct {
	URL        string // Matching URL (aka service name)
	Target     string // Go-Template of target location
	IsTemplate bool   // Is target a go-template?
}

// Rules storage type.
type Storage interface {
	Set(r Rule) error            // add or replace rule
	Get(url string) (Rule, bool) // get location template. should return true if exists
	Remove(url string) error     // remove rule (or ignore if not exists)
	All() ([]*Rule, error)       // dump all save rules
	Reload() error               // reload storage and fill the internal cache
}
