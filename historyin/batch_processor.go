package historyin

import (
	"context"
	"fmt"
	"time"

	"code.justin.tv/foundation/history.v2/internal/batch/processoriface"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/aws/aws-sdk-go/service/kinesis/kinesisiface"
)

type kinesisProcessor struct {
	StreamName  string
	Kinesis     kinesisiface.KinesisAPI
	Logger      Logger
	RunnerState *runnerState
}

func (kp *kinesisProcessor) Process(ctx context.Context, batch []*processoriface.Record) {
	var records []*kinesis.PutRecordsRequestEntry
	for _, rec := range batch {
		if rec == nil {
			continue
		}

		records = append(records, &kinesis.PutRecordsRequestEntry{
			Data:         rec.Data,
			PartitionKey: aws.String(rec.Key),
		})
	}

	kp.sendBatch(ctx, records)
}

func (kp *kinesisProcessor) sendBatch(ctx context.Context, batch []*kinesis.PutRecordsRequestEntry) {
	var nTry int
	for len(batch) > 0 && !kp.RunnerState.Stopped() {
		output, err := kp.Kinesis.PutRecordsWithContext(ctx, &kinesis.PutRecordsInput{
			StreamName: aws.String(kp.StreamName),
			Records:    batch,
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				switch awsErr.Code() {
				case kinesis.ErrCodeProvisionedThroughputExceededException:
					nTry++
					kp.RunnerState.Wait(time.Duration(nTry) * 100 * time.Millisecond)
				}
			}
			kp.Logger.Error(fmt.Errorf("error putting to kinesis: %s", err.Error()))
			continue
		}

		batch, err = kp.failedOnly(batch, output)
		if err != nil {
			kp.Logger.Error(fmt.Errorf("error validating kinesis batch: %s", err.Error()))
			continue
		}

	}
}

func (kp *kinesisProcessor) failedOnly(batch []*kinesis.PutRecordsRequestEntry, output *kinesis.PutRecordsOutput) ([]*kinesis.PutRecordsRequestEntry, error) {
	if len(batch) != len(output.Records) {
		return nil, errInvalidPutBatchResponse
	}

	newBatch := make([]*kinesis.PutRecordsRequestEntry, 0, len(output.Records))
	for nItem, item := range output.Records {
		if item.ErrorCode == nil {
			continue
		}
		// ErrorCodes can be either ProvisionedThroughputExceededException or InternalFailure.
		// Retry in both cases.
		kp.Logger.Error(fmt.Errorf("error sending records to kinesis: %s", aws.StringValue(item.ErrorMessage)))
		newBatch = append(newBatch, batch[nItem])
	}

	return newBatch, nil
}
