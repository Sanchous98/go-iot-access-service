package cache

import (
	"context"
	"github.com/eko/gocache/lib/v4/store"
	"time"
)

const NullCacheType = "nullcache"

type NullStore struct{}

func (n *NullStore) GetWithTTL(context.Context, any) (any, time.Duration, error) {
	return nil, 0, store.NotFound{}
}
func (n *NullStore) Get(context.Context, any) (any, error)                       { return nil, store.NotFound{} }
func (n *NullStore) Set(context.Context, any, any, ...store.Option) error        { return nil }
func (n *NullStore) Delete(context.Context, any) error                           { return nil }
func (n *NullStore) Invalidate(context.Context, ...store.InvalidateOption) error { return nil }
func (n *NullStore) Clear(context.Context) error                                 { return nil }
func (n *NullStore) GetType() string                                             { return NullCacheType }
