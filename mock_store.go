package main

import (
	"reflect"

	"context"

	"google.golang.org/appengine/datastore"
)

type MockStore struct {
	PutFunc              func(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error)
	GetFunc              func(ctx context.Context, key *datastore.Key, dst interface{}) error
	NewQueryFunc         func(kind string) Query
	CacheGetFunc         func(ctx context.Context, key string, dst interface{}) error
	CacheSetFunc         func(ctx context.Context, key string, item interface{}) error
	CacheFlushFunc       func(ctx context.Context) error
	NewKeyFunc           func(ctx context.Context, kind, stringID string, intID int64, parent *datastore.Key) *datastore.Key
	NewIncompleteKeyFunc func(ctx context.Context, kind string, parent *datastore.Key) *datastore.Key
}

func (m *MockStore) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	if m.PutFunc != nil {
		return m.PutFunc(ctx, key, src)
	}
	return key, nil
}

func (m *MockStore) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key, dst)
	}
	return nil
}

func (m *MockStore) NewQuery(kind string) Query {
	if m.NewQueryFunc != nil {
		return m.NewQueryFunc(kind)
	}
	return &MockQuery{}
}

func (m *MockStore) CacheGet(ctx context.Context, key string, dst interface{}) error {
	if m.CacheGetFunc != nil {
		return m.CacheGetFunc(ctx, key, dst)
	}
	return nil
}

func (m *MockStore) CacheSet(ctx context.Context, key string, item interface{}) error {
	if m.CacheSetFunc != nil {
		return m.CacheSetFunc(ctx, key, item)
	}
	return nil
}

func (m *MockStore) CacheFlush(ctx context.Context) error {
	if m.CacheFlushFunc != nil {
		return m.CacheFlushFunc(ctx)
	}
	return nil
}

func (m *MockStore) NewKey(ctx context.Context, kind, stringID string, intID int64, parent *datastore.Key) *datastore.Key {
	if m.NewKeyFunc != nil {
		return m.NewKeyFunc(ctx, kind, stringID, intID, parent)
	}
	// Default behavior: return a dummy key that doesn't panic
	// We can't use datastore.NewKey here if it panics.
	// But datastore.Key is opaque. We can construct one using reflection or just return nil?
	// Returning nil might cause panic in Put/Get if they expect a key.
	// Ideally we return a key that works with our MockStore.
	// Since our MockStore intercepts Put/Get, we can handle nil keys or dummy keys.
	return &datastore.Key{}
}

func (m *MockStore) NewIncompleteKey(ctx context.Context, kind string, parent *datastore.Key) *datastore.Key {
	if m.NewIncompleteKeyFunc != nil {
		return m.NewIncompleteKeyFunc(ctx, kind, parent)
	}
	return &datastore.Key{}
}

type MockQuery struct {
	FilterFunc   func(filterStr string, value interface{}) Query
	OrderFunc    func(fieldName string) Query
	LimitFunc    func(limit int) Query
	AncestorFunc func(ancestor *datastore.Key) Query
	CountFunc    func(ctx context.Context) (int, error)
	GetAllFunc   func(ctx context.Context, dst interface{}) ([]*datastore.Key, error)
}

func (mq *MockQuery) Filter(filterStr string, value interface{}) Query {
	if mq.FilterFunc != nil {
		return mq.FilterFunc(filterStr, value)
	}
	return mq
}

func (mq *MockQuery) Order(fieldName string) Query {
	if mq.OrderFunc != nil {
		return mq.OrderFunc(fieldName)
	}
	return mq
}

func (mq *MockQuery) Limit(limit int) Query {
	if mq.LimitFunc != nil {
		return mq.LimitFunc(limit)
	}
	return mq
}

func (mq *MockQuery) Ancestor(ancestor *datastore.Key) Query {
	if mq.AncestorFunc != nil {
		return mq.AncestorFunc(ancestor)
	}
	return mq
}

func (mq *MockQuery) Count(ctx context.Context) (int, error) {
	if mq.CountFunc != nil {
		return mq.CountFunc(ctx)
	}
	return 0, nil
}

func (mq *MockQuery) GetAll(ctx context.Context, dst interface{}) ([]*datastore.Key, error) {
	if mq.GetAllFunc != nil {
		return mq.GetAllFunc(ctx, dst)
	}
	// Reflection to set dst to empty slice if needed
	v := reflect.ValueOf(dst)
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Slice {
		v.Elem().Set(reflect.MakeSlice(v.Elem().Type(), 0, 0))
	}
	return nil, nil
}
