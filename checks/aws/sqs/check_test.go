package sqs

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

// SimpleMockSQSClient is a manual mock of the SQSAPI interface.
type SimpleMockSQSClient struct {
	// Add fields to store mock outputs and any other state
	GetQueueAttributesOutput *sqs.GetQueueAttributesOutput
	GetQueueAttributesErr    error
}

// GetQueueAttributes is the mock method that mimics the corresponding SQSAPI method.
func (m *SimpleMockSQSClient) GetQueueAttributes(ctx context.Context, params *sqs.GetQueueAttributesInput, optFns ...func(*sqs.Options)) (*sqs.GetQueueAttributesOutput, error) {
	// Return the mocked response and error stored in the mock struct
	return m.GetQueueAttributesOutput, m.GetQueueAttributesErr
}

func TestNew(t *testing.T) {
	queueURL := "http://example.com/queue"
	mockSQSClient := &SimpleMockSQSClient{
		GetQueueAttributesOutput: &sqs.GetQueueAttributesOutput{
			Attributes: map[string]string{
				string(types.QueueAttributeNameQueueArn): "arn:aws:sqs:us-east-1:123456789012:queue1",
			},
		},
		GetQueueAttributesErr: nil,
	}

	config := Config{
		Client:   mockSQSClient,
		QueueUrl: &queueURL,
	}

	checker := New(config)

	err := checker(context.Background())
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
}

func TestNewError(t *testing.T) {
	queueURL := "http://example.com/queue"
	mockSQSClient := &SimpleMockSQSClient{
		GetQueueAttributesOutput: &sqs.GetQueueAttributesOutput{
			Attributes: map[string]string{
				string(types.QueueAttributeNameQueueArn): "arn:aws:sqs:us-east-1:123456789012:queue1",
			},
		},
		GetQueueAttributesErr: errors.New("failed to get queue attributes"),
	}

	config := Config{
		Client:   mockSQSClient,
		QueueUrl: &queueURL,
	}

	checker := New(config)

	err := checker(context.Background())
	if err == nil {
		t.Errorf("Expected error, but got none")
	}
}
