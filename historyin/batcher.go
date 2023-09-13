package historyin

import "time"

// Batcher batches items to kinesis
type Batcher interface {
	Add(audit *Audit) error
	Run()
	Stop(timeout time.Duration) (stopped bool)
	CurrentBatchSize() int
	Drain()
}
