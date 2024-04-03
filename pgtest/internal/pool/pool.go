// Package pool provides a generic resource pool.
package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ResourceConf describes the configuration for dealing with resources owned by
// the pool.
type ResourceConf[T any] struct {
	Create  func(context.Context) (T, error)
	Destroy func(T) error
}

// Pool is a generic resource pool.
//
// As the intended use-case for the pool is to maintain resources used by
// test-cases running in parallel, for simplicity there is currently no maximum
// number of resources owned by the pool. The thinking being that the maximum
// number of resources should already be constrained by the '-parallel' flag
// passed to 'go test' (or GOMAXPROCS by default), so it's not clear if there's
// a use-case where we'd want to set a different limit. If such a use-case
// arrives we can always adjust this definition to allow a maxSize to be
// specified and default to GOMAXPROCS, but for now it is simpler to just let
// it grow as needed.
type Pool[T any] struct {
	resourceConf *ResourceConf[T]

	mut    sync.Mutex
	closed bool
	owned  queue[*Resource[T]]
	idle   stack[*Resource[T]]
}

func New[T any](resourceConf *ResourceConf[T]) *Pool[T] {
	return &Pool[T]{
		resourceConf: resourceConf,
	}
}

func (pool *Pool[T]) createResourceLocked(ctx context.Context) (*Resource[T], error) {
	r, err := pool.resourceConf.Create(ctx)
	if err != nil {
		return nil, err
	}

	resource := &Resource[T]{
		pool: pool,
		data: r,
	}

	pool.owned.enqueue(resource)
	return resource, nil
}

// Acquire acquires the resource from the pool. This can either be a newly
// created resource, or a previously created idle resource.
func (pool *Pool[T]) Acquire(ctx context.Context) (*Resource[T], error) {
	pool.mut.Lock()
	defer pool.mut.Unlock()

	if pool.closed {
		return nil, ErrPoolClosed
	}

	if idleResource, ok := pool.idle.pop(); ok {
		idleResource.state = resourceStateAcquired
		return idleResource, nil
	}

	newResource, err := pool.createResourceLocked(ctx)
	if err != nil {
		return nil, err
	}

	newResource.state = resourceStateAcquired
	return newResource, nil
}

func (pool *Pool[T]) removeOwnedResourceLocked(resource *Resource[T]) {
	if removed := pool.owned.remove(resource); !removed {
		panic(internalVariantBroken("tried to destroy resource not owned by pool"))
	}
}

// Close cleans up the pool and prevents new resources from being acquired.
//
// NOTE: for now we're just destroying ALL resources owned by the pool, not
// just the idle ones. This is primarily for simplicity, since the assumption
// is that pools will only be closed after all tests are done running. As a
// result, we can assume that nobody is using resources acquired from the pool
// when we close the pool. While this is likely not a valid assumption, since
// we probably want to allow for a graceful shutdown in certain cases, it seems
// better to stick with the simple approach for now then adjust as needed.
func (pool *Pool[T]) Close(_ context.Context) error {
	pool.mut.Lock()
	defer pool.mut.Unlock()

	if pool.closed {
		return nil
	}

	pool.closed = true

	var destroyErrs []destroyResourceError[T]

	for {
		toDestroy, ok := pool.owned.dequeue()
		if !ok {
			break
		}

		if err := pool.destroyResourceLocked(toDestroy); err != nil {
			destroyErrs = append(destroyErrs, destroyResourceError[T]{resourceData: toDestroy.data, cause: err})
		}
	}

	if len(destroyErrs) != 0 {
		return destroyResourcesError[T](destroyErrs)
	}

	return nil
}

func (pool *Pool[T]) destroyResourceLocked(resource *Resource[T]) error {
	if err := pool.resourceConf.Destroy(resource.data); err != nil {
		return err
	}

	resource.state = resourceStateDestroyed

	return nil
}

func (pool *Pool[T]) handleResourceReleased(resource *Resource[T]) error {
	pool.mut.Lock()
	defer pool.mut.Unlock()

	if resource.state != resourceStateAcquired {
		return fmt.Errorf("cannot release a non-acquired resource (state=%s)", resource.state)
	}

	// TODO: correctly handle pool closed. For now, since pool.Close
	// destroys all resources and we are holding the lock here, we should
	// never reach this point if the pool has closed (since the resource
	// state will be resourceStateDestroyed). When we update the closing
	// strategy though, which is likely, that will no longer be valid.
	if pool.closed {
		return errors.New("release acquired resource on closed pool")
	}

	resource.state = resourceStateIdle
	pool.idle.push(resource)

	return nil
}

func (pool *Pool[T]) getResourceData(resource *Resource[T]) (data T, err error) {
	pool.mut.Lock()
	defer pool.mut.Unlock()

	if resource.state != resourceStateAcquired && resource.state != resourceStateHijacked {
		return data, fmt.Errorf("cannot get data on a resource that is neither acquired nor hijacked (state=%s)", resource.state)
	}

	return resource.data, nil
}

func (pool *Pool[T]) handleResourceHijacked(resource *Resource[T]) error {
	pool.mut.Lock()
	defer pool.mut.Unlock()

	if resource.state != resourceStateAcquired {
		return fmt.Errorf("cannot hijack a non-acquired resource (state=%s)", resource.state)
	}

	resource.state = resourceStateHijacked
	pool.removeOwnedResourceLocked(resource)

	return nil
}
