package queue

import (
	"sync"

	"github.com/KindCloud97/telegram-bot/model"
	"github.com/google/uuid"
)

type Queue struct {
	list map[string]model.User
	mu   sync.Mutex
}

func NewQueue() *Queue {
	return &Queue{
		list: map[string]model.User{},
		mu:   sync.Mutex{},
	}
}

func (q *Queue) Add(user model.User) {
	q.mu.Lock()
	defer q.mu.Unlock()

	id := uuid.NewString()
	q.list[id] = user
}

// PopUser delete user from list of users.
func (q *Queue) PopUser(id string) (model.User, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	u, ok := q.list[id]
	if !ok {
		return model.User{}, false
	}

	delete(q.list, id)

	return u, ok
}

// GetAll get all users.
func (q *Queue) GetAll() map[string]model.User {
	q.mu.Lock()
	defer q.mu.Unlock()

	allUsers := make(map[string]model.User, len(q.list))
	for k, v := range q.list {
		allUsers[k] = v
	}

	return allUsers
}
