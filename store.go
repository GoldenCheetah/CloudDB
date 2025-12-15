package main

import (
	"context"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/memcache"
)

// Store defines the interface for database and cache operations.
type Store interface {
	Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error)
	Get(ctx context.Context, key *datastore.Key, dst interface{}) error
	NewQuery(kind string) Query

	// Cache operations
	CacheGet(ctx context.Context, key string, dst interface{}) error
	CacheSet(ctx context.Context, key string, item interface{}) error
	CacheFlush(ctx context.Context) error

	// Key operations
	NewKey(ctx context.Context, kind, stringID string, intID int64, parent *datastore.Key) *datastore.Key
	NewIncompleteKey(ctx context.Context, kind string, parent *datastore.Key) *datastore.Key
}

// Query defines the interface for datastore queries.
type Query interface {
	Filter(filterStr string, value interface{}) Query
	Order(fieldName string) Query
	Limit(limit int) Query
	Ancestor(ancestor *datastore.Key) Query
	Count(ctx context.Context) (int, error)
	GetAll(ctx context.Context, dst interface{}) ([]*datastore.Key, error)
}

// DatastoreStore is the concrete implementation of Store using App Engine Datastore.
type DatastoreStore struct{}

func (ds *DatastoreStore) Put(ctx context.Context, key *datastore.Key, src interface{}) (*datastore.Key, error) {
	return datastore.Put(ctx, key, src)
}

func (ds *DatastoreStore) Get(ctx context.Context, key *datastore.Key, dst interface{}) error {
	return datastore.Get(ctx, key, dst)
}

func (ds *DatastoreStore) NewQuery(kind string) Query {
	return &DatastoreQuery{q: datastore.NewQuery(kind)}
}

func (ds *DatastoreStore) CacheGet(ctx context.Context, key string, dst interface{}) error {
	_, err := memcache.Gob.Get(ctx, key, dst)
	return err
}

func (ds *DatastoreStore) CacheSet(ctx context.Context, key string, item interface{}) error {
	return memcache.Gob.Set(ctx, &memcache.Item{
		Key:    key,
		Object: item,
	})
}

func (ds *DatastoreStore) CacheFlush(ctx context.Context) error {
	return memcache.Flush(ctx)
}

func (ds *DatastoreStore) NewKey(ctx context.Context, kind, stringID string, intID int64, parent *datastore.Key) *datastore.Key {
	return datastore.NewKey(ctx, kind, stringID, intID, parent)
}

func (ds *DatastoreStore) NewIncompleteKey(ctx context.Context, kind string, parent *datastore.Key) *datastore.Key {
	return datastore.NewIncompleteKey(ctx, kind, parent)
}

// DatastoreQuery is the concrete implementation of Query.
type DatastoreQuery struct {
	q *datastore.Query
}

func (dq *DatastoreQuery) Filter(filterStr string, value interface{}) Query {
	dq.q = dq.q.Filter(filterStr, value)
	return dq
}

func (dq *DatastoreQuery) Order(fieldName string) Query {
	dq.q = dq.q.Order(fieldName)
	return dq
}

func (dq *DatastoreQuery) Limit(limit int) Query {
	dq.q = dq.q.Limit(limit)
	return dq
}

func (dq *DatastoreQuery) Ancestor(ancestor *datastore.Key) Query {
	dq.q = dq.q.Ancestor(ancestor)
	return dq
}

func (dq *DatastoreQuery) Count(ctx context.Context) (int, error) {
	return dq.q.Count(ctx)
}

func (dq *DatastoreQuery) GetAll(ctx context.Context, dst interface{}) ([]*datastore.Key, error) {
	return dq.q.GetAll(ctx, dst)
}
