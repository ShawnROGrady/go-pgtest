package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
)

type resourceCounts struct {
	created   atomic.Uint32
	destroyed atomic.Uint32
}

func (counts *resourceCounts) assertCreatedEqualsDestroyed(t testing.TB) bool {
	t.Helper()
	created, destroyed := counts.created.Load(), counts.destroyed.Load()
	if created != destroyed {
		t.Errorf("created=%d; destroyed=%d", created, destroyed)
		return false
	}
	return true
}

func countedResourceConf[T any](counts *resourceCounts, conf *ResourceConf[T]) *ResourceConf[T] {
	return &ResourceConf[T]{
		Create: func(ctx context.Context) (T, error) {
			x, err := conf.Create(ctx)
			if err == nil {
				counts.created.Add(1)
			}
			return x, err
		},
		Destroy: func(x T) error {
			err := conf.Destroy(x)
			if err == nil {
				counts.destroyed.Add(1)
			}
			return err
		},
	}
}

type fakeResource struct {
	mut    sync.Mutex
	closed bool
}

func (r *fakeResource) Close() {
	r.mut.Lock()
	defer r.mut.Unlock()
	r.closed = true
}

func (r *fakeResource) Valid() error {
	r.mut.Lock()
	defer r.mut.Unlock()
	if r.closed {
		return errors.New("fakeResource closed")
	}
	return nil
}

var fakeResourceConf = &ResourceConf[*fakeResource]{
	Create: func(ctx context.Context) (*fakeResource, error) {
		return new(fakeResource), nil
	},
	Destroy: func(x *fakeResource) error {
		x.Close()
		return nil
	},
}

func TestPoolSingleWorker(t *testing.T) {
	worker := func(ctx context.Context, pool *Pool[*fakeResource]) error {
		r, err := pool.Acquire(ctx)
		if err != nil {
			return err
		}
		defer r.Release()

		return r.Data().Valid()
	}

	var (
		ctx    = context.Background()
		counts = new(resourceCounts)
	)

	pool := New(countedResourceConf(counts, fakeResourceConf))

	if err := worker(ctx, pool); err != nil {
		t.Fatalf("worker: %s", err)
	}

	if err := pool.Close(ctx); err != nil {
		t.Fatalf("close: %s", err)
	}

	counts.assertCreatedEqualsDestroyed(t)
}

func TestPoolAcquireThenAcquire(t *testing.T) {
	ctx := context.Background()
	pool := New(fakeResourceConf)

	if _, err := pool.Acquire(ctx); err != nil {
		t.Fatalf("first acquire: %s", err)
	}
	if _, err := pool.Acquire(ctx); err != nil {
		t.Fatalf("second acquire: %s", err)
	}

	t.Logf("pool.owned.items = %v", pool.owned.items())
}
