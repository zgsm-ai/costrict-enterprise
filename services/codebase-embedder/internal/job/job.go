package job

import "context"

type Job interface {
	Start()
	Close()
}

type Processor interface {
	Process(ctx context.Context) error
}
