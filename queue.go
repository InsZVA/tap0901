package tap0901

const QUEUE_SIZE = 1024

type queue struct {
	data [QUEUE_SIZE]interface{}
	read int
	write int
}

func (q *queue) pop(t *interface{}) bool {
	if q.data[q.read] == nil {
		return false
	}
	*t = q.data[q.read]
	q.data[q.read] = nil
	if q.read + 1 >= QUEUE_SIZE {
		q.read += 1 - QUEUE_SIZE
	} else {
		q.read++
	}
	return true
}

func (q *queue) push(t interface{}) bool {
	if q.data[q.write] == nil {
		// WMB
		q.data[q.write] = t
		if q.write + 1 >= QUEUE_SIZE {
			q.write += 1 - QUEUE_SIZE
		} else {
			q.write++
		}
		return true
	}
	return false
}