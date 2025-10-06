package queue

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
)

type Handlers struct{ Logger *slog.Logger }

func (h *Handlers) NotifyUserHandler(c context.Context, t *asynq.Task) error {
	var p NotifyUserPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}
	h.Logger.Info("notify user", "user_id", p.UserID, "message", p.Message)
	// TODO: kirim email/push/notifikasi di sini
	return nil
}

func NewServer(redisOpt asynq.RedisClientOpt, concurrency int, logger *slog.Logger) *asynq.Server {
	return asynq.NewServer(redisOpt, asynq.Config{Concurrency: concurrency, Queues: map[string]int{"critical": 2, "default": 8}})
}

func NewMux(h *Handlers) *asynq.ServeMux {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TypeNotifyUser, h.NotifyUserHandler)
	return mux
}
