package historyin

import (
	"testing"
	"time"

	"code.justin.tv/foundation/history.v2/mocks"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type BatchRunnerKinesisSuite struct {
	suite.Suite
	mockKinesis *mocks.KinesisAPI
	batchRunner *batchRunner
}

func (s *BatchRunnerKinesisSuite) SetupTest() {
	s.mockKinesis = new(mocks.KinesisAPI)
	rs := new(runnerState)
	processor := &kinesisProcessor{
		StreamName:  "test-data-stream-name",
		Kinesis:     s.mockKinesis,
		RunnerState: rs,
		Logger:      nopLogger{},
	}

	s.batchRunner = &batchRunner{
		RunnerState:    rs,
		BatchProcessor: processor,
	}
}

func (s *BatchRunnerKinesisSuite) TearDownTest() {
	s.mockKinesis.AssertExpectations(s.T())
}

func (s *BatchRunnerKinesisSuite) TestRunSingleLoop() {
	s.Require().NoError(s.batchRunner.Add(&Audit{}))
	s.mockKinesis.
		On("PutRecordsWithContext", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			input := args.Get(1).(*kinesis.PutRecordsInput)
			s.Assert().Len(input.Records, 1)
			s.Assert().False(
				s.batchRunner.Stop(time.Duration(time.Duration(0))),
				"expect able to be stopped since this is in Run")
		}).
		Return(&kinesis.PutRecordsOutput{}, nil)

	go s.batchRunner.Run()
	s.Assert().True(s.batchRunner.RunnerState.Wait(time.Second))
}

func TestBatchRunner(t *testing.T) {
	suite.Run(t, &BatchRunnerKinesisSuite{})
}
