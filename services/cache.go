package services

import (
	"sync"
	"time"
)

// CacheData almacena los datos en caché junto con el timestamp y el estado de actualización.
type CacheData[T any] struct {
	Data       T
	Timestamp  time.Time
	IsUpdating bool
}

// CacheService maneja la lógica de almacenamiento en caché.
type CacheService[T any] struct {
	cache *CacheData[T]
	ttl   time.Duration
	mu    sync.Mutex
}

// NewCacheService crea una nueva instancia de CacheService con el TTL (en minutos).
func NewCacheService[T any](ttlInMinutes int) *CacheService[T] {
	return &CacheService[T]{
		ttl: time.Duration(ttlInMinutes) * time.Minute,
	}
}

// Set almacena nuevos datos en caché.
func (c *CacheService[T]) Set(data T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = &CacheData[T]{
		Data:       data,
		Timestamp:  time.Now(),
		IsUpdating: false,
	}
}

// Get retorna los datos en caché si existen y no han expirado.
func (c *CacheService[T]) Get() (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache != nil && time.Since(c.cache.Timestamp) < c.ttl {
		return c.cache.Data, true
	}
	var empty T
	return empty, false
}

// NeedsUpdate indica si la caché ha expirado y necesita actualizarse.
func (c *CacheService[T]) NeedsUpdate() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cache == nil {
		return true
	}
	return time.Since(c.cache.Timestamp) > c.ttl && !c.cache.IsUpdating
}

// SetUpdating marca el estado de actualización de la caché.
func (c *CacheService[T]) SetUpdating(status bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cache != nil {
		c.cache.IsUpdating = status
	}
}
