// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package inmem

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/armon/go-radix"
	log "github.com/hashicorp/go-hclog"
	"github.com/openbao/openbao/api/v2"
	"github.com/openbao/openbao/sdk/v2/physical"
)

// Verify interfaces are satisfied
var (
	_ physical.Backend   = (*InmemBackend)(nil)
	_ physical.HABackend = (*InmemHABackend)(nil)
	_ physical.Lock      = (*InmemLock)(nil)
)

var (
	PutDisabledError      = errors.New("put operations disabled in inmem backend")
	GetDisabledError      = errors.New("get operations disabled in inmem backend")
	DeleteDisabledError   = errors.New("delete operations disabled in inmem backend")
	ListDisabledError     = errors.New("list operations disabled in inmem backend")
	GetInTxnDisabledError = errors.New("get operations inside transactions are disabled in inmem backend")
)

// InmemBackend is an in-memory only physical backend. It is useful
// for testing and development situations where the data is not
// expected to be durable.
type InmemBackend struct {
	sync.RWMutex
	root         *radix.Tree
	permitPool   *physical.PermitPool
	logger       log.Logger
	failGet      *uint32
	failPut      *uint32
	failDelete   *uint32
	failList     *uint32
	failGetInTxn *uint32
	logOps       bool
	maxValueSize int
}

// NewInmem constructs a new in-memory backend
func NewInmem(conf map[string]string, logger log.Logger) (physical.Backend, error) {
	maxValueSize := 0
	maxValueSizeStr, ok := conf["max_value_size"]
	if ok {
		var err error
		maxValueSize, err = strconv.Atoi(maxValueSizeStr)
		if err != nil {
			return nil, err
		}
	}

	return &InmemBackend{
		root:         radix.New(),
		permitPool:   physical.NewPermitPool(physical.DefaultParallelOperations),
		logger:       logger,
		failGet:      new(uint32),
		failPut:      new(uint32),
		failDelete:   new(uint32),
		failList:     new(uint32),
		failGetInTxn: new(uint32),
		logOps:       api.ReadBaoVariable("BAO_INMEM_LOG_ALL_OPS") != "",
		maxValueSize: maxValueSize,
	}, nil
}

// Put is used to insert or update an entry
func (i *InmemBackend) Put(ctx context.Context, entry *physical.Entry) error {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.Lock()
	defer i.Unlock()

	return i.PutInternal(ctx, entry)
}

func (i *InmemBackend) PutInternal(ctx context.Context, entry *physical.Entry) error {
	if i.logOps {
		i.logger.Trace("put", "key", entry.Key)
	}
	if atomic.LoadUint32(i.failPut) != 0 {
		return PutDisabledError
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if i.maxValueSize > 0 && len(entry.Value) > i.maxValueSize {
		return fmt.Errorf("%s", physical.ErrValueTooLarge)
	}

	i.root.Insert(entry.Key, entry.Value)
	return nil
}

func (i *InmemBackend) FailPut(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failPut, val)
}

// Get is used to fetch an entry
func (i *InmemBackend) Get(ctx context.Context, key string) (*physical.Entry, error) {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.RLock()
	defer i.RUnlock()

	return i.GetInternal(ctx, key)
}

func (i *InmemBackend) GetInternal(ctx context.Context, key string) (*physical.Entry, error) {
	if i.logOps {
		i.logger.Trace("get", "key", key)
	}
	if atomic.LoadUint32(i.failGet) != 0 {
		return nil, GetDisabledError
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if raw, ok := i.root.Get(key); ok {
		return &physical.Entry{
			Key:   key,
			Value: raw.([]byte),
		}, nil
	}
	return nil, nil
}

func (i *InmemBackend) FailGet(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failGet, val)
}

func (i *InmemBackend) FailGetInTxn(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failGetInTxn, val)
}

// Delete is used to permanently delete an entry
func (i *InmemBackend) Delete(ctx context.Context, key string) error {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.Lock()
	defer i.Unlock()

	return i.DeleteInternal(ctx, key)
}

func (i *InmemBackend) DeleteInternal(ctx context.Context, key string) error {
	if i.logOps {
		i.logger.Trace("delete", "key", key)
	}
	if atomic.LoadUint32(i.failDelete) != 0 {
		return DeleteDisabledError
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	i.root.Delete(key)
	return nil
}

func (i *InmemBackend) FailDelete(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failDelete, val)
}

// List is used to list all the keys under a given
// prefix, up to the next prefix.
func (i *InmemBackend) List(ctx context.Context, prefix string) ([]string, error) {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.RLock()
	defer i.RUnlock()

	return i.ListInternal(ctx, prefix)
}

// ListPage is used to list all the keys under a given
// prefix, up to the next prefix, but limiting to a
// specified number of keys after a given entry.
func (i *InmemBackend) ListPage(ctx context.Context, prefix string, after string, limit int) ([]string, error) {
	i.permitPool.Acquire()
	defer i.permitPool.Release()

	i.RLock()
	defer i.RUnlock()

	return i.ListPaginatedInternal(ctx, prefix, after, limit)
}

func (i *InmemBackend) ListInternal(ctx context.Context, prefix string) ([]string, error) {
	return i.ListPaginatedInternal(ctx, prefix, "", -1)
}

func (i *InmemBackend) ListPaginatedInternal(ctx context.Context, prefix string, after string, limit int) ([]string, error) {
	if i.logOps {
		i.logger.Trace("list", "prefix", prefix)
	}
	if atomic.LoadUint32(i.failList) != 0 {
		return nil, ListDisabledError
	}

	var out []string
	seen := make(map[string]interface{})
	walkFn := func(s string, v interface{}) bool {
		if limit > 0 && len(out) >= limit {
			// We've seen enough entries; exit early.
			return true
		}

		// Note that we push the comparison with trimmed down until
		// after we add in the directory suffix, if necessary.
		trimmed := strings.TrimPrefix(s, prefix)
		sep := strings.Index(trimmed, "/")
		if sep == -1 {
			if after != "" && trimmed <= after {
				// Still prior to our cut-off point, so retry.
				return false
			}

			out = append(out, trimmed)
		} else {
			// Include the directory suffix to distinguish keys from
			// subtrees.
			trimmed = trimmed[:sep+1]
			if after != "" && trimmed <= after {
				// Still prior to our cut-off point, so retry.
				return false
			}

			if _, ok := seen[trimmed]; !ok {
				out = append(out, trimmed)
				seen[trimmed] = struct{}{}
			}
		}

		return false
	}
	i.root.WalkPrefix(prefix, walkFn)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return out, nil
}

func (i *InmemBackend) FailList(fail bool) {
	var val uint32
	if fail {
		val = 1
	}
	atomic.StoreUint32(i.failList, val)
}
