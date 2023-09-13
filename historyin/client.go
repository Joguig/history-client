package historyin

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"code.justin.tv/foundation/history.v2/internal/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

// size when batch will be flushed to firehose
const flushBatchSize = 250

// batch age when batch will be flushed to firehose
const flushBatchAge = time.Minute

// Client adds history events
type Client struct {
	// Environment is the history stack to use. Defaults to production.
	Environment    string
	FlushBatchSize int
	Logger         Logger

	initSync sync.Once

	streamName string
	kinesis    kinesisiface.KinesisAPI
}

func (c *Client) init() (err error) {
	c.initSync.Do(func() {
		var cfg config.Config
		if cfg, err = config.Environment(c.Environment); err != nil {
			return
		}

		var kinesisSession *session.Session
		kinesisSession, err = session.NewSession(&aws.Config{Region: aws.String(cfg.AWSRegion)})
		if err != nil {
			return
		}

		kinesisSession, err = session.NewSession(&aws.Config{
			Region:      aws.String(cfg.AWSRegion),
			Credentials: stscreds.NewCredentials(kinesisSession, cfg.RoleARN),
		})
		if err != nil {
			return
		}

		c.streamName = cfg.StreamName
		c.kinesis = kinesis.New(kinesisSession)

		if c.Logger == nil {
			c.Logger = nopLogger{}
		}

		if c.FlushBatchSize == 0 {
			c.FlushBatchSize = flushBatchSize
		}

		if c.FlushBatchSize > kinesisBatchMaxRecords {
			err = fmt.Errorf("FlushBatchSize must be less than %d", c.FlushBatchSize)
			return
		}
	})

	return
}

// Add submits a new audit to history service
func (c *Client) Add(ctx context.Context, audit *Audit) error {
	if err := c.init(); err != nil {
		return err
	}

	if err := audit.fillOptional(); err != nil {
		return err
	}

	partitionKey := string(audit.UUID)
	data, err := json.Marshal(audit)
	if err != nil {
		return err
	}

	if _, err := c.kinesis.PutRecordWithContext(ctx, &kinesis.PutRecordInput{
		Data:         data,
		PartitionKey: aws.String(partitionKey),
		StreamName:   aws.String(c.streamName),
	}); err != nil {
		return err
	}

	return nil
}

// Batcher returns a new batcher
func (c *Client) Batcher() (Batcher, error) {
	if err := c.init(); err != nil {
		return nil, err
	}

	rs := new(runnerState)
	processor := &kinesisProcessor{
		StreamName:  c.streamName,
		Kinesis:     c.kinesis,
		RunnerState: rs,
		Logger:      c.Logger,
	}

	return &batchRunner{
		Batch: batch{
			Threshold: flushBatchSize,
		},
		MaxBatchAge:    flushBatchAge,
		RunnerState:    rs,
		BatchProcessor: processor,
	}, nil
}
