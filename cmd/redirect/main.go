package main

import (
	_ "embed"
	"flag"
	"net"
	"net/http"
	"time"

	"github.com/reddec/redirect"
	"github.com/sirupsen/logrus"
)

func main() {
	runUI := flag.Bool("ui", false, "run ui")
	uiFolder := flag.String("ui-files", "", "Location of custom UI files")
	uiAddr := flag.String("ui-addr", "127.0.0.1:10101", "Address for UI")
	configFile := flag.String("config", "./redir.json", "File to save configs")
	pollInterval := flag.String("poll", "5s", "Polling interval")
	bind := flag.String("bind", "0.0.0.0:10100", "Redirect address")
	flag.Parse()

	// get redirect port for UI
	_, port, _ := net.SplitHostPort(*bind)

	// init defaults
	stats := redirect.InMemoryStats()
	storage := &redirect.JSONStorage{FileName: *configFile}
	if err := storage.Reload(); err != nil {
		logrus.Fatal(err)
	}

	engine := redirect.DefaultEngine(storage, stats)
	if err := engine.Reload(); err != nil {
		logrus.Fatal(err)
	}

	interval, err := time.ParseDuration(*pollInterval)
	if err != nil {
		logrus.Fatal(err)
	}

	watcher := redirect.NewWatcher(interval)
	watcher.AddFile(*configFile, func() {
		if err := storage.Reload(); err != nil {
			logrus.Warning("failed to reload config: %s", err)
			return
		}

		if err := engine.Reload(); err != nil {
			logrus.Warning("failed to reload engine: %s", err)
			return
		}
		logrus.Info("config reloaded")
	})
	watcher.Watch()

	if *runUI {
		ui := redirect.DefaultUI(storage, stats, engine, port)
		go startUI(*uiAddr, *uiFolder, ui)
	}

	logrus.Info("Bind:", *bind)
	panic(http.ListenAndServe(*bind, redirect.WithLogging(engine)))
}

func startUI(uiAddr string, uiFolder string, ui http.Handler) {
	static := http.FileServer(http.FS(redirect.DefaultUIStatic()))
	if uiFolder != "" {
		static = http.FileServer(http.Dir(uiFolder))
	}
	http.Handle("/ui/", redirect.WithLogging(static))
	http.Handle("/api/", redirect.WithLogging(http.StripPrefix("/api/", ui)))
	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		// redirect to ui
		http.Redirect(writer, request, "ui/", http.StatusTemporaryRedirect)
	})
	logrus.Info("UI:", uiAddr)
	panic(http.ListenAndServe(uiAddr, nil))
}
