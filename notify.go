package redirect

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type monitor struct {
	filename     string
	lastModified time.Time
	callback     func()
}

type Watcher struct {
	interval time.Duration
	ticker   *time.Ticker
	stop     chan bool
	monitor  []*monitor
}

func NewWatcher(i time.Duration) *Watcher {
	return &Watcher{
		interval: i,
		ticker:   time.NewTicker(i),
		monitor:  []*monitor{},
	}
}

func (w *Watcher) Watch() {
	go w.watch()
}

func (w *Watcher) watch() {
	for {
		select {
		case <-w.stop:
			return
		case <-w.ticker.C:
			w.checkFiles()
		}
	}
}

func (w *Watcher) Stop() {
	w.ticker.Stop()
	w.stop <- true
}

func (w *Watcher) AddFile(filename string, callback func()) error {
	mtime, err := w.getMtime(filename)
	if err != nil {
		return err
	}
	m := &monitor{
		filename:     filename,
		lastModified: mtime,
		callback:     callback,
	}
	w.monitor = append(w.monitor, m)
	return nil
}

func (w *Watcher) getMtime(filename string) (time.Time, error) {
	fi, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}
	return fi.ModTime(), nil
}

func (w *Watcher) checkFiles() {
	for _, m := range w.monitor {
		mtime, err := w.getMtime(m.filename)
		if err != nil {
			continue
		}
		logrus.WithFields(logrus.Fields{
			"filename": m.filename,
			"mtime":    mtime,
			"last":     m.lastModified,
		}).Debug("check file")
		if !mtime.After(m.lastModified) {
			continue
		}
		m.lastModified = mtime
		m.callback()
	}
}
