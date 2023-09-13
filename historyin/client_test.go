package historyin

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"code.justin.tv/foundation/history.v2/mocks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var (
	testCreatedAt = time.Now()
)

type ClientSuite struct {
	suite.Suite

	mockKinesis *mocks.KinesisAPI
	client      *Client
}

func (s *ClientSuite) SetupTest() {
	s.mockKinesis = new(mocks.KinesisAPI)
	s.client = &Client{
		streamName: s.streamName(),
		kinesis:    s.mockKinesis,
	}
	s.client.initSync.Do(func() {})
}

func (s *ClientSuite) TearDownTest() {
	s.mockKinesis.AssertExpectations(s.T())
}

func (s *ClientSuite) streamName() string {
	return "my-kinesis-data-stream-name"
}

func (s *ClientSuite) dummyAudit() *Audit {
	return &Audit{
		UUID:         "myuuidmyuuidmyuuidmyuuidmyuuidmyuuid",
		Action:       "my-action",
		UserType:     "my-user-type",
		UserID:       "my-user-id",
		ResourceType: "my-resource-type",
		ResourceID:   "my-resource-id",
		Description:  "my-description",
		CreatedAt:    Time(testCreatedAt),
		TTL:          Duration(time.Hour),
		Changes: []ChangeSet{
			{
				Attribute: "cs-attribute",
				OldValue:  "cs-old-value",
				NewValue:  "cs-new-value",
			},
		},
	}
}

func (s *ClientSuite) dummyAuditMarshalled() []byte {
	data, err := json.Marshal(s.dummyAudit())
	s.Require().NoError(err)
	return data
}

func (s *ClientSuite) TestValidEnv() {
	var client *Client
	for _, env := range []string{"", "staging", "prod"} {
		client = &Client{
			Environment: env,
		}
		s.Require().NoError(client.init())
		s.Assert().NotEmpty(client.streamName)
		s.Assert().NotEmpty(client.kinesis)
		s.Assert().NotEmpty(client.FlushBatchSize)
	}
}

func (s *ClientSuite) TestInitInvalidEnv() {
	client := &Client{
		Environment: "invalid",
	}
	s.Assert().Error(client.init())
}

func (s *ClientSuite) TestInitInvalidFlushBatchSize() {
	client := &Client{
		Environment:    "prod",
		FlushBatchSize: 501,
	}

	err := client.init()
	s.Require().Error(err)
	s.Assert().True(strings.HasPrefix(err.Error(), "FlushBatchSize must be"))
}

func (s *ClientSuite) mockDummyKinesisPut() *mock.Call {
	return s.mockKinesis.
		On("PutRecordWithContext", mock.Anything, &kinesis.PutRecordInput{
			Data:         s.dummyAuditMarshalled(),
			StreamName:   aws.String(s.streamName()),
			PartitionKey: aws.String(string(s.dummyAudit().UUID)),
		})
}

func (s *ClientSuite) TestAddSuccess() {
	s.mockDummyKinesisPut().
		Return(nil, nil)

	s.Require().NoError(s.client.Add(context.Background(), s.dummyAudit()))
}

func (s *ClientSuite) TestAddSuccessEmptyUUID() {
	dummyAudit := s.dummyAudit()
	dummyAudit.UUID = ""
	s.mockKinesis.
		On("PutRecordWithContext", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			input := args.Get(1).(*kinesis.PutRecordInput)
			a := new(audit)
			s.Require().NoError(json.Unmarshal(input.Data, a))
			s.Assert().NotEmpty(a.UUID)
		}).
		Return(nil, nil)

	s.Require().NoError(s.client.Add(context.Background(), dummyAudit))
}

func (s *ClientSuite) TestAddUUIDError() {
	s.Assert().Error(s.client.Add(context.Background(), &Audit{
		UUID: "bad-uuid",
	}))
}

func (s *ClientSuite) TestAddAWSError() {
	myErr := errors.New("my-error")
	s.mockDummyKinesisPut().
		Return(nil, myErr)

	s.Assert().Equal(myErr, s.client.Add(context.Background(), s.dummyAudit()))
}

func (s *ClientSuite) TestBatcher() {
	b, err := s.client.Batcher()
	s.Assert().NoError(err)
	batcher := b.(*batchRunner)
	s.Assert().IsType(&kinesisProcessor{}, batcher.BatchProcessor)
	bp := batcher.BatchProcessor.(*kinesisProcessor)
	s.Assert().Equal(batcher.RunnerState, bp.RunnerState)
}

func TestClient(t *testing.T) {
	suite.Run(t, &ClientSuite{})
}
