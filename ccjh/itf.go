package ccjh

import (
	"context"
	"time"
)

type JobMeta interface {
	SetResult(any)
	SetError(error)
	SetStarted(time.Time)
	SetCompleted(time.Time)
	IsCanceled() bool
}

type TargetFunc func(context.Context, map[string]any) (any, error)
