package pool

import "fmt"

// resourceState describes the state of a resource that has been created by the
// pool.
type resourceState uint

const (
	// The resource has been created. The main purpose of this variant, and
	// the reason it is the default value, is to allow us to create
	// resource before defining ownership.
	resourceStateCreated resourceState = iota

	// The resource has been created and is available to be acquired.
	resourceStateIdle

	// The resource has been acquired and is being used.
	resourceStateAcquired

	// Ownership of the resource has been transferred from the pool, and
	// the new owner is now responsible for maintaining it.
	resourceStateHijacked

	// The resource has been destroyed.
	resourceStateDestroyed
)

func (state resourceState) String() string {
	//exhaustive:enforce
	switch state {
	case resourceStateCreated:
		return "created"
	case resourceStateIdle:
		return "idle"
	case resourceStateAcquired:
		return "acquired"
	case resourceStateHijacked:
		return "hijacked"
	case resourceStateDestroyed:
		return "destroyed"
	}

	panic(internalVariantBroken(fmt.Sprintf("invalid resourceState: %d", state)))
}

// A Resource is a resource maintained by a Pool.
type Resource[T any] struct {
	pool  *Pool[T]
	data  T
	state resourceState
}

// Release releases the resource back to the pool so it can be re-used.
func (r *Resource[T]) Release() {
	if err := r.pool.handleResourceReleased(r); err != nil {
		panic(internalVariantBroken(err.Error()))
	}
}

// Data returns the underlying data associated with the resource.
func (r *Resource[T]) Data() T {
	d, err := r.pool.getResourceData(r)
	if err != nil {
		panic(internalVariantBroken(err.Error()))
	}

	return d
}

// Hijack takes control of the resource from the pool.
func (r *Resource[T]) Hijack() {
	if err := r.pool.handleResourceHijacked(r); err != nil {
		panic(internalVariantBroken(err.Error()))
	}
}
