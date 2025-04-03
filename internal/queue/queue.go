package queue

import "sync"

type AlertEvent struct {
	Host        string
	Description string
	Type        string // "ALERT" 或 "RECOVERY"
}

type Queue struct {
	mu    sync.Mutex
	items []AlertEvent
}

func NewQueue() *Queue {
	return &Queue{
		items: make([]AlertEvent, 0),
	}
}

func (q *Queue) Push(event AlertEvent) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.items = append(q.items, event)
}

func (q *Queue) Pop() (AlertEvent, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) == 0 {
		return AlertEvent{}, false
	}
	event := q.items[0]
	q.items = q.items[1:]
	return event, true
}
