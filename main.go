package resource

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Const list
const (
	DefaultExpiration time.Duration = 0
)

// Error list
var (
	ErrNotFound = errors.New("not found")
)

// Coding Task: Concurrent in-memory cache.
//
// Fetcher (see below) is an interface which abstracts the process of fetching
// and loading a "Model".  In practice there would be Fetcher implementations
// for retrieving and loading models from local file systems, S3 buckets etc...
//
// Implement and test an in-memory cache which wraps a given Fetcher and caches
// calls to its Fetch method (complete the implementation of NewCache and the
// FetchCache type below).

// Model is a resource.
type Model struct {
	Name string
}

// Fetcher is an interface that defines the Fetch method.
type Fetcher interface {
	// Fetch retrieves an Model for a given identifier id.
	Fetch(ctx context.Context, id string) (*Model, error)
}

// NewCache creates a new Fetcher which caches calls to f.Fetch.
// See FetchCache for more details.
func NewCache(f Fetcher) *FetchCache {
	return &FetchCache{
		cache: newCache(),
		f:     f,
	}
}

func newCache() *cache {
	return &cache{
		items: make(map[string]item),
	}
}

// FetchCache implements an in-memory cache for a Fetcher.
//
// A FetchCache is safe for use by multiple goroutines simultaneously.
type FetchCache struct {
	f Fetcher
	*cache
}

type cache struct {
	items map[string]item
	mu    sync.RWMutex
}

// item --
type item struct {
	Object     *Model
	Expiration int64
}

// expired Returns true if the item has expired.
func (item item) expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// Fetch implements Fetcher.
func (fc *FetchCache) Fetch(ctx context.Context, id string) (*Model, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	item, found := fc.fetchFromCache(id)
	if !found {
		return fc.fetchFromFetcher(ctx, id)
	}

	return item.Object, nil
}

// Clear item by id
func (fc *FetchCache) Clear(id string) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	if _, found := fc.items[id]; !found {
		return
	}

	delete(fc.items, id)
}

func (fc *FetchCache) fetchFromCache(id string) (item, bool) {
	i, found := fc.items[id]
	if !found || i.expired() {
		return item{}, false
	}

	return i, found
}

func (fc *FetchCache) fetchFromFetcher(ctx context.Context, id string) (*Model, error) {
	model, err := fc.f.Fetch(ctx, id)
	if err != nil {
		return nil, err
	}

	fc.cacheitem(id, model)

	return model, nil
}

func (fc *FetchCache) cacheitem(id string, model *Model) {
	fc.items[id] = item{
		Object:     model,
		Expiration: int64(DefaultExpiration),
	}
}
