package queue

import (
	"github.com/hibiken/asynq"
)

func NewScheduler(redisOpt asynq.RedisClientOpt) *asynq.Scheduler {
	return asynq.NewScheduler(redisOpt, &asynq.SchedulerOpts{})
}

func RegisterSchedules(s *asynq.Scheduler) error {
	// contoh: push heartbeat tiap menit
	_, err := s.Register("* * * * *", asynq.NewTask("heartbeat", nil))
	return err
}
