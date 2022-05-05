package redirect

import (
	"embed"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

const (
	formFieldTarget     = "target"
	formFieldIsTemplate = "is_template"
	formFieldService    = "service"
	headerRedirPort     = "X-Redir-Port"
)

//go:embed ui/*
var defaultStaticUI embed.FS // nolint:gochecknoglobals

var stripProtocol = regexp.MustCompile(`^(?:https?://)`)

// Get default UI static files (prefixed by ui/).
func DefaultUIStatic() embed.FS {
	return defaultStaticUI
}

// description of rule for API request.
type UIEntry struct {
	Target     string `json:"target"`
	Hits       int64  `json:"hits"`
	URL        string `json:"url"`
	IsTemplate bool   `json:"isTemplate"`
}

type basicUI struct {
	storage   Storage
	stats     StatReader
	engine    Engine
	redirPort string
}

func DefaultUI(storage Storage, stats StatReader, engine Engine, redirPort string) http.Handler {
	if storage == nil {
		panic("ui storage is nil")
	}
	if stats == nil {
		panic("ui stats reader is nil")
	}
	if engine == nil {
		panic("ui engine ref is nil")
	}
	return &basicUI{
		stats:     stats,
		storage:   storage,
		engine:    engine,
		redirPort: redirPort,
	}
}

func (ui *basicUI) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	defer rq.Body.Close()
	service := strings.Trim(rq.URL.Path, "/")
	switch rq.Method {
	case http.MethodGet:
		if service == "" {
			ui.list(wr, rq)
		} else {
			ui.get(service, wr, rq)
		}
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		ui.set(wr, rq)
	case http.MethodDelete:
		ui.remove(service, wr, rq)
	default:
		ui.list(wr, rq)
	}
}

func (ui *basicUI) list(wr http.ResponseWriter, _ *http.Request) {
	var ans = make(map[string]*UIEntry)
	entries, err := ui.storage.All()
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, elem := range entries {
		ans[elem.URL] = &UIEntry{
			URL:        elem.URL,
			Target:     elem.Target,
			IsTemplate: elem.IsTemplate,
			Hits:       ui.stats.Visits(elem.URL),
		}
	}
	wr.Header().Set(headerRedirPort, ui.redirPort)
	sendJSON(ans, wr)
}

func (ui *basicUI) get(service string, wr http.ResponseWriter, rq *http.Request) {
	rule, exists := ui.storage.Get(service)
	if !exists {
		http.NotFound(wr, rq)
		return
	}
	wr.Header().Set(headerRedirPort, ui.redirPort)
	sendJSON(&UIEntry{
		URL:        service,
		Hits:       ui.stats.Visits(service),
		Target:     rule.Target,
		IsTemplate: rule.IsTemplate,
	}, wr)
}

func (ui *basicUI) remove(service string, wr http.ResponseWriter, _ *http.Request) {
	err := ui.storage.Remove(service)
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	err = ui.engine.Reload()
	if err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	wr.WriteHeader(http.StatusNoContent)
}

func (ui *basicUI) set(wr http.ResponseWriter, rq *http.Request) {
	var entry UIEntry
	if strings.Contains(rq.Header.Get("Content-Type"), "application/json") {
		// parse entry as-is except hits
		err := json.NewDecoder(rq.Body).Decode(&entry)
		if err != nil {
			http.Error(wr, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		// use form
		err := rq.ParseForm()
		if err != nil {
			http.Error(wr, err.Error(), http.StatusBadRequest)
			return
		}
		entry.URL = rq.FormValue(formFieldService)
		entry.Target = rq.FormValue(formFieldTarget)
		entry.IsTemplate = strings.ToLower(rq.FormValue(formFieldIsTemplate)) == "true"
	}
	url := stripProtocol.ReplaceAllString(entry.URL, "")
	target := stripProtocol.ReplaceAllString(entry.Target, "")
	r := Rule{
		URL:        url,
		Target:     target,
		IsTemplate: entry.IsTemplate,
	}
	if err := ui.storage.Set(r); err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := ui.engine.Reload(); err != nil {
		http.Error(wr, err.Error(), http.StatusInternalServerError)
		return
	}
	wr.WriteHeader(http.StatusNoContent)
}

// correctly send JSON with required headers.
func sendJSON(data interface{}, w http.ResponseWriter) {
	content, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}
