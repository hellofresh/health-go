package sqs

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// Config is the AWS SQS checker configuration settings.
type Config struct {
	// Client is the initialized instance of the AWS SQS client used for making API requests to the AWS SQS service.
	Client SqsActions
	// QueueUrl is a pointer to the string that contains the URL of the SQS queue to which the client connects.
	QueueUrl *string
}

// SqsActions defines the set of operations used from the SQS client.
type SqsActions interface {
	GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error)
}

// New creates a new AWS SQS health check that verifies if a connection to the queue exists
func New(config Config) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		arnAttributeName := types.QueueAttributeNameQueueArn
		if _, err := config.Client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl:       config.QueueUrl,
			AttributeNames: []types.QueueAttributeName{arnAttributeName},
		}); err != nil {
			return fmt.Errorf("unable to get queue ARN: %w", err)
		}

		return nil
	}
}
