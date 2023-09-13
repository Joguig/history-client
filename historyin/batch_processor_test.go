package historyin

import (
	"testing"

	"code.justin.tv/foundation/history.v2/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/suite"
)

type KinesisProcessorSuite struct {
	suite.Suite
	mockKinesis *mocks.KinesisAPI
	processor   *kinesisProcessor
}

func (s *KinesisProcessorSuite) SetupTest() {
	s.mockKinesis = new(mocks.KinesisAPI)
	s.processor = &kinesisProcessor{
		StreamName:  "mock",
		RunnerState: new(runnerState),
		Logger:      nopLogger{},
		Kinesis:     s.mockKinesis,
	}
}

func (s *KinesisProcessorSuite) TearDownTest() {
	s.mockKinesis.AssertExpectations(s.T())
}

func (s *KinesisProcessorSuite) TestFailedOnlyEmptyBatch() {
	fb, err := s.processor.failedOnly(
		[]*kinesis.PutRecordsRequestEntry{},
		&kinesis.PutRecordsOutput{
			Records: []*kinesis.PutRecordsResultEntry{},
		})

	s.Require().NoError(err)
	s.Assert().Empty(fb)
}

func (s *KinesisProcessorSuite) TestFailedOnlyDifferentLens() {
	_, err := s.processor.failedOnly(
		[]*kinesis.PutRecordsRequestEntry{
			{},
		},
		&kinesis.PutRecordsOutput{
			Records: []*kinesis.PutRecordsResultEntry{},
		})

	s.Assert().Equal(errInvalidPutBatchResponse, err)
}

func (s *KinesisProcessorSuite) TestFailedOnlyAllFailed() {
	records := []*kinesis.PutRecordsRequestEntry{{}}
	fb, err := s.processor.failedOnly(
		records,
		&kinesis.PutRecordsOutput{
			Records: []*kinesis.PutRecordsResultEntry{
				{ErrorCode: aws.String("error-code")},
			},
		})

	s.Assert().NoError(err)
	s.Assert().Equal(records, fb)
}

func (s *KinesisProcessorSuite) TestFailedOnlyNoFailed() {
	fb, err := s.processor.failedOnly(
		[]*kinesis.PutRecordsRequestEntry{{}},
		&kinesis.PutRecordsOutput{
			Records: []*kinesis.PutRecordsResultEntry{
				{},
			},
		})

	s.Assert().NoError(err)
	s.Assert().Empty(fb)
}

func TestBatchProcessor(t *testing.T) {
	suite.Run(t, &KinesisProcessorSuite{})
}
