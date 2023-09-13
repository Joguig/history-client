package processoriface

import "context"

// Record to process by batch processor
type Record struct {
	Key  string
	Data []byte
}

// ProcessorAPI .....
type ProcessorAPI interface {
	Process(ctx context.Context, batch []*Record)
}
