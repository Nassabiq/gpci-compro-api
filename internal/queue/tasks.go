package queue

import (
	"encoding/json"

	"github.com/hibiken/asynq"
)

const (
	TypeNotifyUser = "notify:user"
)

type NotifyUserPayload struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

func NewNotifyUserTask(p NotifyUserPayload) (*asynq.Task, error) {
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(TypeNotifyUser, b), nil
}

func EnqueueNotifyUser(client *asynq.Client, payload NotifyUserPayload) (*asynq.TaskInfo, error) {
	task, err := NewNotifyUserTask(payload)
	if err != nil {
		return nil, err
	}
	return client.Enqueue(task, asynq.Queue("default"))
}
