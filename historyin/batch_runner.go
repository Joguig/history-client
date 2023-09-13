package historyin

import (
	"context"
	"errors"
	"time"

	"code.justin.tv/foundation/history.v2/internal/batch/processoriface"
)

const (
	kinesisBatchMaxRecords = 500
)

var (
	errInvalidPutBatchResponse = errors.New("invalid put batch response")
)

// BatchRunner sends batches to kinesis
type batchRunner struct {
	MaxBatchAge    time.Duration
	Batch          batch
	BatchProcessor processoriface.ProcessorAPI
	RunnerState    *runnerState
}

// Add adds an audit to the batch
func (br *batchRunner) Add(audit *Audit) error {
	return br.Batch.Add(audit)
}

// BatchSize returns the size of the batch
func (br *batchRunner) CurrentBatchSize() int {
	return br.Batch.CurrentSize()
}

// Run pushes batches to kinesis
func (br *batchRunner) Run() {
	for !br.RunnerState.Stopped() {
		ctx := br.RunnerState.Context()
		br.waitForWork(ctx)
		batch := br.Batch.PopBatch(kinesisBatchMaxRecords)
		if len(batch) > 0 {
			br.BatchProcessor.Process(ctx, batch)
		}
	}

	br.RunnerState.MarkDone()
}

func (br *batchRunner) waitForWork(ctx context.Context) {
	if br.RunnerState.IsDraining() {
		return
	}

	select {
	case <-ctx.Done():
	case <-time.NewTimer(br.MaxBatchAge).C:
	case <-br.Batch.ThresholdBreach():
		br.Batch.MarkThresholdBreachRead()
	}
}

func (br *batchRunner) Drain() {
	br.RunnerState.Drain()
}

// Stop the Runner
func (br *batchRunner) Stop(timeout time.Duration) (stopped bool) {
	br.RunnerState.Stop()
	return br.RunnerState.Wait(timeout)
}
