package mocks

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

type MockEventPublisher struct {
	mock.Mock
}

func NewMockEventPublisher() *MockEventPublisher {
	return &MockEventPublisher{}
}

// ! implements -->
func (m *MockEventPublisher) GetChannel() (*amqp.Channel, error) {
	return nil, nil
	// args := m.Called()
	// return args.Get(0), args.Error(1)
}

func (m *MockEventPublisher) Close() error {
	args := m.Called()
	return args.Error(1)
}

func (m *MockEventPublisher) DeclareExchange(name, kind string) error {
	args := m.Called(name, kind)
	return args.Error(0)
}

func (m *MockEventPublisher) Publish(ctx context.Context, exchange, routingKey string, options amqp.Publishing) error {
	args := m.Called("Publish", ctx, exchange, routingKey, options)
	return args.Error(1)
}

func (m *MockEventPublisher) CreateQueue(queueName string, durable, autodelete bool) (amqp.Queue, error) {
	args := m.Called("CreateQueue", queueName, durable, autodelete)
	return args.Get(0).(amqp.Queue), args.Error(1)
}

func (m *MockEventPublisher) CreateBinding(name, binding, exchange string) error {
	args := m.Called("CreateBinding", name, binding, exchange)
	return args.Error(0)
}
func (m *MockEventPublisher) Consume(queueName, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
	args := m.Called(queueName, consumer, autoAck)
	return args.Get(0).(chan amqp.Delivery), args.Error(1)
}
