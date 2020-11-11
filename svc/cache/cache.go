package cache

import (
	"time"
)

type (
	Item struct {
		Item interface{}
		Code string
	}

	ICacheCache interface {
		Get(key string) (*Item, bool)
		Replace(key string, value *Item, d time.Duration) error
		ReplaceIfExists(key string, value *Item, d time.Duration) error
		Delete(key string)
		Add(key string, value *Item, d time.Duration) error

		IsEnabled() bool
		IsDisabled() bool

		Flush()
		Start()
		Shutdown()
	}
)
