mockery := mockery -all -note 'Code generated DO NOT EDIT'

.PHONY: mocks
dev:
	fresh -c fresh.conf
mocks:
	$(mockery) -dir vendor/github.com/aws/aws-sdk-go/service/kinesis/kinesisiface
