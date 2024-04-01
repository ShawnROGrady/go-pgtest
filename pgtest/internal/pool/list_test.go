package pool

import (
	"fmt"
	"slices"
	"testing"
)

func TestQueueEnqueue(t *testing.T) {
	q := new(queue[string])
	q.enqueue("first")
	q.enqueue("second")
	q.enqueue("third")

	t.Logf("q = %#v", q)
	if q.empty() {
		t.Error("after enqueues q.empty() = true; want false")
	}

	expectedItems := []string{"first", "second", "third"}
	actualItems := q.items()

	if !slices.Equal(actualItems, expectedItems) {
		t.Errorf("q.items() = %q; want %q", actualItems, expectedItems)
	}
}

type expectedDequeue[T any] struct {
	empty     bool
	popped    T
	remaining []T
}

func checkDequeue[T comparable](t testing.TB, prefix string, q *queue[T], expected expectedDequeue[T]) {
	t.Helper()

	popped, ok := q.dequeue()
	expectedOk := !expected.empty

	if (ok != expectedOk) || (popped != expected.popped) {
		t.Errorf(
			"%s q.dequeue() = (%v, %t); want (%v, %t)", prefix,
			popped, ok,
			expected.popped, expectedOk,
		)
	}

	remaining := q.items()
	if !slices.Equal(remaining, expected.remaining) {
		t.Errorf(
			"%s q.items() = %v; want %v", prefix,
			remaining, expected.remaining,
		)
	}
}

func TestQueueDequeue(t *testing.T) {
	var q queue[string]
	q.enqueue("foo")
	q.enqueue("bar")
	q.enqueue("baz")

	checkDequeue(t, "first", &q, expectedDequeue[string]{
		popped:    "foo",
		remaining: []string{"bar", "baz"},
	})

	checkDequeue(t, "second", &q, expectedDequeue[string]{
		popped:    "bar",
		remaining: []string{"baz"},
	})

	checkDequeue(t, "third", &q, expectedDequeue[string]{
		popped:    "baz",
		remaining: nil,
	})

	checkDequeue(t, "fourth", &q, expectedDequeue[string]{
		empty:     true,
		remaining: nil,
	})

	if !q.empty() {
		t.Error("after fourth pop q.empty() = true; want false")
	}
}

func TestQueueRemoves(t *testing.T) {
	testCases := map[string]struct {
		initialItems       []string
		toRemove           string
		expectRemoved      bool
		expectedFinalItems []string
	}{
		"empty_queue": {
			initialItems:  nil,
			toRemove:      "foo",
			expectRemoved: false,
		},
		"first_item": {
			initialItems:       []string{"foo", "bar", "baz"},
			toRemove:           "foo",
			expectRemoved:      true,
			expectedFinalItems: []string{"bar", "baz"},
		},
		"last_item": {
			initialItems:       []string{"foo", "bar", "baz"},
			toRemove:           "baz",
			expectRemoved:      true,
			expectedFinalItems: []string{"foo", "bar"},
		},
		"middle_item": {
			initialItems:       []string{"foo", "bar", "baz"},
			toRemove:           "bar",
			expectRemoved:      true,
			expectedFinalItems: []string{"foo", "baz"},
		},
		"item_not_in_queue": {
			initialItems:  []string{"foo", "bar", "baz"},
			toRemove:      "qux",
			expectRemoved: false,
		},
	}

	for testName, testCase := range testCases {
		t.Run(testName, func(t *testing.T) {
			t.Logf("initialItems = %q", testCase.initialItems)

			q := new(queue[string])
			for _, initialItem := range testCase.initialItems {
				q.enqueue(initialItem)
			}

			removed := q.remove(testCase.toRemove)
			if removed != testCase.expectRemoved {
				t.Errorf("q.remove(%q) = %t; want %t", testCase.toRemove, removed, testCase.expectRemoved)
			}

			finalItems := q.items()
			expectedFinalItems := testCase.expectedFinalItems
			if !testCase.expectRemoved {
				expectedFinalItems = testCase.initialItems
			}

			if !slices.Equal(finalItems, expectedFinalItems) {
				t.Errorf("after q.remove(%q) q.items() = %q; want %q", testCase.toRemove, finalItems, expectedFinalItems)
			}
		})
	}
}

type qContainer[T comparable] struct{ q queue[T] }

func TestQueueDequeueAll(t *testing.T) {
	container := new(qContainer[string])

	container.q.enqueue("first")
	container.q.enqueue("second")
	container.q.enqueue("third")

	var items []string
	for {
		item, ok := container.q.dequeue()
		if !ok {
			break
		}

		items = append(items, item)
	}

	expectedItems := []string{"first", "second", "third"}
	if !slices.Equal(items, expectedItems) {
		t.Errorf("items = %q; want %q", items, expectedItems)
	}
}

type user struct {
	name string
}

func newUser(name string) *user {
	return &user{name: name}
}

func (u *user) GoString() string {
	if u == nil {
		return "nil"
	}
	return fmt.Sprintf("newUser(%q)", u.name)
}

func TestQueueDequeueAllPointer(t *testing.T) {
	toEnqueue := []*user{
		newUser("first"),
		newUser("second"),
		newUser("third"),
	}

	container := new(qContainer[*user])
	for _, x := range toEnqueue {
		container.q.enqueue(x)
	}

	var items []*user
	for {
		item, ok := container.q.dequeue()
		if !ok {
			break
		}

		items = append(items, item)
	}

	expectedItems := slices.Clone(toEnqueue)
	if !slices.Equal(items, expectedItems) {
		t.Errorf("items = %v; want %v", items, expectedItems)
	}
}

func TestStackPush(t *testing.T) {
	st := new(stack[string])
	st.push("first")
	st.push("second")
	st.push("third")

	if st.empty() {
		t.Error("after pushes st.empty() = true; want false")
	}

	expectedItems := []string{"third", "second", "first"}
	actualItems := st.items()

	if !slices.Equal(actualItems, expectedItems) {
		t.Errorf("st.items() = %q; want %q", actualItems, expectedItems)
	}
}

type expectedPop[T any] struct {
	empty     bool
	popped    T
	remaining []T
}

func checkPop[T comparable](t testing.TB, prefix string, st *stack[T], expected expectedPop[T]) {
	t.Helper()

	popped, ok := st.pop()
	expectedOk := !expected.empty

	if (ok != expectedOk) || (popped != expected.popped) {
		t.Errorf(
			"%s st.dequeue() = (%v, %t); want (%v, %t)", prefix,
			popped, ok,
			expected.popped, expectedOk,
		)
	}

	remaining := st.items()
	if !slices.Equal(remaining, expected.remaining) {
		t.Errorf(
			"%s st.items() = %v; want %v", prefix,
			remaining, expected.remaining,
		)
	}
}

func TestStackPop(t *testing.T) {
	st := new(stack[string])
	st.push("foo")
	st.push("bar")
	st.push("baz")

	checkPop(t, "first", st, expectedPop[string]{
		popped:    "baz",
		remaining: []string{"bar", "foo"},
	})

	checkPop(t, "second", st, expectedPop[string]{
		popped:    "bar",
		remaining: []string{"foo"},
	})

	checkPop(t, "third", st, expectedPop[string]{
		popped:    "foo",
		remaining: nil,
	})

	checkPop(t, "fourth", st, expectedPop[string]{
		empty:     true,
		remaining: nil,
	})

	if !st.empty() {
		t.Error("after fourth pop st.empty() = true; want false")
	}
}
