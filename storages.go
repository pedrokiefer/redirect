package redirect

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

// Simple single-file storage. All rules saved as-is by JSON indented encoder to the provided file after each Set ops.
type JSONStorage struct {
	FileName string // File name to store and read
	cache    map[string]Rule
	lock     sync.RWMutex
}

type storage struct {
	Rules []Rule `json:"rules"`
}

// Set or replace one rule, serialize cache to JSON and then dump to disk. Even if dump failed rule is saved into cache.
func (js *JSONStorage) Set(r Rule) error {
	js.lock.Lock()
	defer js.lock.Unlock()
	if js.cache == nil {
		js.cache = make(map[string]Rule)
	}
	js.cache[r.URL] = r
	return js.unsafeDump()
}

// Get single record from cache.
func (js *JSONStorage) Get(url string) (Rule, bool) {
	js.lock.RLock()
	defer js.lock.RUnlock()
	v, ok := js.cache[url]
	return v, ok
}

// Remove rule from cache and save dump to disk. Even if dump failed rule removed from cache.
func (js *JSONStorage) Remove(url string) error {
	if js.cache == nil {
		return nil
	}
	js.lock.Lock()
	defer js.lock.Unlock()
	delete(js.cache, url)
	return js.unsafeDump()
}

// All rules stored in cache. Never returns error.
func (js *JSONStorage) All() ([]*Rule, error) {
	var ans = make([]*Rule, 0, len(js.cache))
	js.lock.RLock()
	defer js.lock.RUnlock()
	for _, r := range js.cache {
		c := r
		ans = append(ans, &c)
	}
	return ans, nil
}

// Read all rules from file. Will not update cache if file will not exists.
func (js *JSONStorage) Reload() error {
	js.lock.RLock() // prevent read and write the same file
	data, err := ioutil.ReadFile(js.FileName)
	js.lock.RUnlock()
	if err != nil {
		if os.IsNotExist(err) {
			log.Println(js.FileName, err)
			return nil
		}
		return fmt.Errorf("read JSON config: %w", err)
	}

	var list storage
	err = json.Unmarshal(data, &list)
	if err != nil {
		return fmt.Errorf("parse JSON config: %w", err)
	}

	js.UpdateCache(list.Rules)
	return nil
}

func (js *JSONStorage) UpdateCache(rules []Rule) {
	cache := make(map[string]Rule, len(rules))
	for _, r := range rules {
		cache[r.URL] = r
	}
	js.lock.Lock()
	js.cache = cache
	js.lock.Unlock()
}

func (js *JSONStorage) unsafeDump() error {
	list := storage{}
	list.Rules = make([]Rule, 0, len(js.cache))
	for _, r := range js.cache {
		list.Rules = append(list.Rules, r)
	}
	data, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("marshal JSON config: %w", err)
	}
	return ioutil.WriteFile(js.FileName, data, 0600)
}
