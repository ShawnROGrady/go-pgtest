package pool

import "fmt"

type listNode[T any] struct {
	val T
	tl  *listNode[T]
}

func (node *listNode[T]) GoString() string {
	if node == nil {
		return "nil"
	}

	return fmt.Sprintf("&listNode{val: %#v, tl: %#v}", node.val, node.tl)
}

func (node *listNode[T]) pushBack(v T) {
	if node.tl == nil {
		node.tl = &listNode[T]{val: v}
		return
	}

	node.tl.pushBack(v)
}

func (node *listNode[T]) items() []T {
	var (
		vs  []T
		cur = node
	)

	for cur != nil {
		vs = append(vs, cur.val)
		cur = cur.tl
	}

	return vs
}

type queue[T comparable] struct {
	hd *listNode[T]
}

func (q *queue[T]) empty() bool {
	return q.hd == nil
}

func (q *queue[T]) enqueue(v T) {
	if q.hd == nil {
		q.hd = &listNode[T]{val: v}
		return
	}

	q.hd.pushBack(v)
}

func (q *queue[T]) dequeue() (T, bool) {
	if q.hd == nil {
		var nothing T
		return nothing, false
	}

	dequeued := q.hd.val
	q.hd = q.hd.tl
	return dequeued, true
}

func (q *queue[T]) remove(v T) bool {
	var (
		prev = q.hd
		cur  = prev
	)

	for cur != nil {
		if cur.val == v {
			break
		}

		prev = cur
		cur = cur.tl
	}

	if cur == nil {
		return false
	}

	if cur == q.hd {
		q.hd = cur.tl
		return true
	}

	prev.tl = cur.tl
	cur.tl = nil

	return true
}

func (q *queue[T]) items() []T {
	return q.hd.items()
}

func (q queue[T]) GoString() string {
	return fmt.Sprintf("&queue{hd: %#v}", q.hd)
}

type stack[T any] struct {
	hd *listNode[T]
}

func (st *stack[T]) empty() bool {
	return st.hd == nil
}

func (st *stack[T]) push(v T) {
	st.hd = &listNode[T]{
		val: v,
		tl:  st.hd,
	}
}

func (st *stack[T]) pop() (T, bool) {
	if st.hd == nil {
		var nothing T
		return nothing, false
	}

	popped := st.hd.val
	st.hd = st.hd.tl
	return popped, true
}

func (st *stack[T]) items() []T {
	return st.hd.items()
}
