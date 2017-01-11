package tap0901

import "testing"

var q *queue

func assert(v bool, t *testing.T, errStr string) {
	if !v {
		t.Error(errStr)
	}
}

func Test_queue_push(t *testing.T) {
	q = &queue{}
	assert(q.push(1), t, "push failed!")
	assert(q.push(2), t, "push failed!")
	assert(q.push(3), t, "push failed!")
	assert(q.push(4), t, "push failed!")
	assert(q.push(5), t, "push failed!")
	for i := 5; i < QUEUE_SIZE; i++ {
		q.push(i+1)
	}
	assert(!q.push(1), t, "push failed, not full!")
}

func Test_queue_pop(t *testing.T) {
	var i interface{}
	q.pop(&i)
	v, ok := i.(int)
	assert(ok, t, "pop failed, type error")
	assert(v == 1, t, "pop failed, number error")
	q.pop(&i)
	v, ok = i.(int)
	assert(ok, t, "pop failed, type error")
	assert(v == 2, t, "pop failed, number error")
	q2 := queue{}
	assert(!q2.pop(&i), t, "pop failed, not emtpy!")
}