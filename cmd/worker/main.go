package main

import (
	"os"

	"github.com/Nassabiq/gpci-compro-api/internal/config"
	"github.com/Nassabiq/gpci-compro-api/internal/queue"
	"github.com/Nassabiq/gpci-compro-api/internal/utils"
	"github.com/hibiken/asynq"
)

func main() {
	cfg := config.Load()
	logger := utils.NewLogger(cfg.App.Env)
	redisOpt := asynq.RedisClientOpt{Addr: cfg.Redis.Addr, Password: cfg.Redis.Password, DB: cfg.Redis.DB}
	server := queue.NewServer(redisOpt, cfg.Asynq.Concurrency, logger)
	mux := queue.NewMux(&queue.Handlers{Logger: logger})
	logger.Info("worker started")
	if err := server.Run(mux); err != nil {
		logger.Error("worker exit", "err", err)
		os.Exit(1)
	}
}
