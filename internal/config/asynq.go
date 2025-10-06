package config

type AsynqConfig struct {
	Concurrency   int
	QueueDefault  string
	QueueCritical string
}

func loadAsynqConfig() AsynqConfig {
	return AsynqConfig{
		Concurrency:   mustInt("ASYNQ_CONCURRENCY", 10),
		QueueDefault:  getenv("ASYNQ_QUEUE_DEFAULT", "default"),
		QueueCritical: getenv("ASYNQ_QUEUE_CRITICAL", "critical"),
	}
}
