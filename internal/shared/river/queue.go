package riverqueue

import (
	"time"

	"github.com/riverqueue/river"
)

type QueueConfig struct {
	MaxWorkers      int           `json:"max_workers"`
	PollInterval    time.Duration `json:"poll_interval"`
	RescueStuckJobs time.Duration `json:"rescue_stuck_jobs"`
	RetryPolicy     *RetryPolicy  `json:"retry_policy"`
}

// RetryPolicy defines the retry behavior for jobs in the queue.
type RetryPolicy struct {
	MaxAttempts  int           `json:"max_attempts"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
}

// DefaultQueueConfig returns the default queue configuration
func DefaultQueueConfig() *QueueConfig {
	return &QueueConfig{
		MaxWorkers:      10,
		PollInterval:    1 * time.Second,
		RescueStuckJobs: 5 * time.Minute,
		RetryPolicy: &RetryPolicy{
			MaxAttempts:  3,
			InitialDelay: 1 * time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
		},
	}
}

// GetQueues returns the queue configuration map
func (c *QueueConfig) GetQueues() map[string]river.QueueConfig {
	return map[string]river.QueueConfig{
		river.QueueDefault: {MaxWorkers: c.MaxWorkers},
		"high_priority":    {MaxWorkers: c.MaxWorkers * 2},
		"low_priority":     {MaxWorkers: max(1, c.MaxWorkers/2)},
		"notifications":    {MaxWorkers: 5},
		"emails":           {MaxWorkers: 3},
		"moderation":       {MaxWorkers: 4},
		"webhooks":         {MaxWorkers: 2},
		"background":       {MaxWorkers: 2},
		"analytics":        {MaxWorkers: 2},
		"token":            {MaxWorkers: 2},
		"cleanup":          {MaxWorkers: 1},
	}
}

type QueueName string

const (
	DefaultQueue       QueueName = "default"
	HighPriorityQueue  QueueName = "high_priority"
	LowPriorityQueue   QueueName = "low_priority"
	NotificationsQueue QueueName = "notifications"
	EmailsQueue        QueueName = "emails"
	ModerationQueue    QueueName = "moderation"
	WebhooksQueue      QueueName = "webhooks"
	BackgroundQueue    QueueName = "background"
	AnalyticsQueue     QueueName = "analytics"
	CleanupQueue       QueueName = "cleanup"
)
