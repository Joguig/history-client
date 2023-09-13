package historyin

import (
	"encoding/json"
	"sync"

	"code.justin.tv/foundation/history.v2/internal/batch/processoriface"
)

// batch batches audits to send to kinesis
type batch struct {
	Threshold int

	initSync    sync.Once
	recordsLock sync.Mutex
	records     []*processoriface.Record

	thresholdBreachOnce *sync.Once
	thresholdBreachLock sync.Mutex
	thresholdBreach     chan struct{}
}

func (b *batch) init() {
	b.initSync.Do(func() {
		b.thresholdBreach = make(chan struct{}, 1)
		if b.Threshold == 0 {
			b.Threshold = 1
		}

		b.thresholdBreachOnce = new(sync.Once)
	})
}

// Add adds a record to the batch
func (b *batch) Add(audit *Audit) error {
	b.init()

	if err := audit.fillOptional(); err != nil {
		return err
	}

	data, err := json.Marshal(audit)
	if err != nil {
		return err
	}

	b.recordsLock.Lock()
	defer b.recordsLock.Unlock()

	b.records = append(b.records, &processoriface.Record{
		Data: data,
		Key:  string(audit.UUID),
	})

	if len(b.records) >= b.Threshold {
		b.getThresholdBreachOnce().Do(func() {
			b.thresholdBreach <- struct{}{}
		})
	}

	return nil
}

func (b *batch) CurrentSize() int {
	b.init()

	b.recordsLock.Lock()
	defer b.recordsLock.Unlock()
	return len(b.records)
}

// PopBatch pops a batch to send to kinesis
func (b *batch) PopBatch(maxSize int) []*processoriface.Record {
	b.init()

	b.recordsLock.Lock()
	defer b.recordsLock.Unlock()

	if len(b.records) == 0 {
		return nil
	}

	if len(b.records) <= maxSize {
		records := b.records
		b.records = []*processoriface.Record{}
		return records
	}

	records := b.records[0:maxSize]
	if len(b.records) > maxSize {
		b.records = b.records[maxSize:len(b.records)]
	} else {
		b.records = []*processoriface.Record{}
	}

	return records
}

func (b *batch) ThresholdBreach() <-chan struct{} {
	b.init()

	return b.thresholdBreach
}

func (b *batch) ThresholdBreached() bool {
	b.init()

	return b.CurrentSize() >= b.Threshold
}

func (b *batch) MarkThresholdBreachRead() {
	b.init()

	b.thresholdBreachLock.Lock()
	b.thresholdBreachOnce = new(sync.Once)
	b.thresholdBreachLock.Unlock()
}

func (b *batch) getThresholdBreachOnce() *sync.Once {
	b.thresholdBreachLock.Lock()
	lock := b.thresholdBreachOnce
	b.thresholdBreachLock.Unlock()
	return lock
}
