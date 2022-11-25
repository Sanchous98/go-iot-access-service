package cache

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/utils"
	"context"
	"errors"
	"fmt"
	"github.com/eko/gocache/v3/store"
	"github.com/goccy/go-json"
	"log"
)

type Cache[T any] struct {
	store store.StoreInterface `inject:""`
}

func New[T any](store store.StoreInterface) *Cache[T] {
	return &Cache[T]{store}
}

func (c *Cache[T]) Get(ctx context.Context, key any) (*T, error) {
	item, err := c.store.Get(ctx, key)

	if err != nil {
		if !errors.Is(err, new(store.NotFound)) {
			log.Println(err)
		}
		return new(T), err
	}

	var result T

	err = json.UnmarshalNoEscape(utils.StrToBytes(item.(string)), &result)
	return &result, err
}

func (c *Cache[T]) Set(ctx context.Context, key any, object *T, options ...store.Option) error {
	item, err := json.MarshalNoEscape(object)

	switch key.(type) {
	case []byte:
		key = utils.BytesToStr(key.([]byte))
	case string:
	case fmt.Stringer:
		key = key.(fmt.Stringer).String()
	}

	if err != nil {
		return err
	}

	return c.store.Set(ctx, key, item, options...)
}

func (c *Cache[T]) Delete(ctx context.Context, key any) error {
	return c.store.Delete(ctx, key)
}

func (c *Cache[T]) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return c.store.Invalidate(ctx, options...)
}

func (c *Cache[T]) Clear(ctx context.Context) error {
	return c.store.Clear(ctx)
}

func (c *Cache[T]) GetType() string {
	return c.store.GetType()
}
