package redirect

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"text/template"

	"github.com/sirupsen/logrus"
)

type redirect struct {
	t      *template.Template
	target string
}

type engine struct {
	storage Storage
	stat    StatWriter
	lock    sync.RWMutex
	rules   map[string]redirect
}

// Create default engine based on provided storage and sink.
func DefaultEngine(storage Storage, sink StatWriter) Engine {
	if storage == nil {
		panic("storage is nil")
	}
	if sink == nil {
		panic("stats sink is nil")
	}
	return &engine{
		storage: storage,
		stat:    sink,
	}
}

func (eng *engine) ServeHTTP(wr http.ResponseWriter, rq *http.Request) {
	defer rq.Body.Close()

	service := rq.Host

	// try to find redirect rule
	eng.lock.RLock()
	r, ok := eng.rules[service]
	eng.lock.RUnlock()
	if !ok {
		http.NotFound(wr, rq)
		return
	}
	// notify stat counter
	eng.stat.Touch(service)

	url := r.target
	if r.t != nil {
		// render redirect template
		urlData := &bytes.Buffer{}
		err := r.t.Execute(urlData, rq)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"service": service,
			}).Error("failed execute template for service", err)
			http.Error(wr, err.Error(), http.StatusInternalServerError)
			return
		}
		url = strings.TrimSpace(urlData.String())
	}

	wr.Header().Add("Content-Length", "0")
	http.Redirect(wr, rq, "https://"+url, http.StatusMovedPermanently)
}

func (eng *engine) Reload() error {
	rules, err := eng.storage.All()
	if err != nil {
		return fmt.Errorf("engine: read rules from storage: %w", err)
	}
	var swap = make(map[string]redirect, len(rules))
	for _, rule := range rules {
		var tpl *template.Template
		if rule.IsTemplate {
			tpl, err = template.New("").Parse(rule.Target)
			if err != nil {
				return fmt.Errorf("engine: parse rule for url %v: %w", rule.URL, err)
			}
		}
		swap[rule.URL] = redirect{
			t:      tpl,
			target: rule.Target,
		}
	}
	eng.lock.Lock()
	eng.rules = swap
	eng.lock.Unlock()
	return nil
}
