package job

import (
	"context"

	"github.com/zgsm-ai/codebase-indexer/internal/svc"
)

type Scheduler struct {
	Jobs []Job
}

func NewScheduler(serverCtx context.Context, svcCtx *svc.ServiceContext) (*Scheduler, error) {
	cleaner, err := NewCleaner(serverCtx, svcCtx)
	if err != nil {
		return nil, err
	}
	jobs := []Job{
		cleaner,
	}
	return &Scheduler{
		Jobs: jobs,
	}, nil
}

func (s *Scheduler) Schedule() {
	for _, job := range s.Jobs {
		go job.Start()
	}
}

func (s *Scheduler) Close() {
	for _, job := range s.Jobs {
		if job == nil {
			continue
		}
		job.Close()
	}
}
