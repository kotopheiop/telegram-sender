package message

import "sync"

type Message struct {
	ChatID          int64
	MessageThreadID *int
	Text            string
}

type MessageQueue struct {
	sync.Mutex
	messages []Message
}

func (q *MessageQueue) AddToTheBeginningEnqueue(message Message) {
	q.Lock()
	defer q.Unlock()
	q.messages = append([]Message{message}, q.messages...)
}

func (q *MessageQueue) Enqueue(message Message) {
	q.Lock()
	defer q.Unlock()
	q.messages = append(q.messages, message)
}

func (q *MessageQueue) Dequeue() *Message {
	q.Lock()
	defer q.Unlock()
	if len(q.messages) == 0 {
		return nil
	}
	message := q.messages[0]
	q.messages = q.messages[1:]
	return &message
}
