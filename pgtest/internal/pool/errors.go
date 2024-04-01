package pool

import (
	"errors"
	"fmt"
	"strings"
)

var ErrPoolClosed = errors.New("pgtest: pool closed")

// An internalVariantBrokenError indicates that some internal variant of this
// library has been broken. These indicate something that should be impossible
// happenned, so we should panic instead of returning an error, so the main
// reason for this type is to provide consistent messaging in these situations.
type internalVariantBrokenError struct {
	description string
}

func (e internalVariantBrokenError) Error() string {
	return "pgtest: internal variant broken - " + e.description
}

func internalVariantBroken(description string) error {
	return internalVariantBrokenError{description: description}
}

type destroyResourceError[T any] struct {
	resourceData T
	cause        error
}

func (e destroyResourceError[T]) Error() string {
	return fmt.Sprintf("pgtest: destroy resource %v: %s", e.resourceData, e.cause)
}

type destroyResourcesError[T any] []destroyResourceError[T]

func (e destroyResourcesError[T]) Unwrap() []error {
	errs := make([]error, len(e))
	for i, err := range e {
		errs[i] = err
	}

	return errs
}

func (e destroyResourcesError[T]) Error() string {
	parts := make([]string, len(e))
	for i, err := range e {
		parts[i] = fmt.Sprintf("(%v: %s)", err.resourceData, err.cause)
	}

	return "pgtest: destroy resources: [" + strings.Join(parts, ", ") + "]"
}
