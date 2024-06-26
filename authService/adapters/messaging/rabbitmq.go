package messaging

// import (
// 	"auth-svc/internal/ports"
// 	"fmt"

// 	amqp "github.com/rabbitmq/amqp091-go"
// 	//"github.com/rabbitmq/amqp091-go/amqp"
// 	// "github.com/streadway/amqp" // Import the streadway/amqp library
// )

// type RabbitMQClient struct {
// 	config ports.Config
// 	conn   *amqp.Connection
// }

// func NewRabbitClient(config ports.Config) (*RabbitMQClient, error) {
// 	cfg := config.GetBrokerConfig()
// 	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.User, cfg.Password, cfg.Host, cfg.Port)

// 	conn, err := amqp.Dial(dsn)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
// 	}

// 	return &RabbitMQClient{config: config, conn: conn}, nil
// }

// func (rc *RabbitMQClient) GetChannel() (*amqp.Channel, error) {
// 	ch, err := rc.conn.Channel()
// 	if err != nil {
// 		return nil, err
// 	}

// 	// if err := ch.Confirm(false); err != nil {
// 	// 	return nil, nil
// 	// }

// 	return ch, nil
// }

// func (rc *RabbitMQClient) Close() error {
// 	return rc.conn.Close() // Close the underlying connection
// }

// // CreateExchange declares a new exchange on the RabbitMQ server
// func (rc *RabbitMQClient) DeclareExchange(name, kind string) error {
// 	ch, err := rc.GetChannel()
// 	if err != nil {
// 		return err
// 	}
// 	defer ch.Close() // Close the channel after use

// 	return ch.ExchangeDeclare(
// 		name,  // Name of the exchange
// 		kind,  // Type of exchange (e.g., "fanout", "direct", "topic")
// 		true,  // Durable (survives server restarts)
// 		false, // Delete when unused
// 		false, // Exclusive (only this connection can access)
// 		false,
// 		nil, // Arguments
// 	)
// }

// // CreateQueue declares a new queue on the RabbitMQ server
// func (rc *RabbitMQClient) CreateQueue(queueName string, durable, autodelete bool) (amqp.Queue, error) {
// 	ch, err := rc.GetChannel()
// 	if err != nil {
// 		return amqp.Queue{}, err
// 	}
// 	defer ch.Close() // Close the channel after use

// 	queue, err := ch.QueueDeclare(
// 		queueName,  // Name of the queue
// 		durable,    // Durable (survives server restarts)
// 		autodelete, // Exclusive (only this connection can access)
// 		false,      // Delete when unused
// 		false,
// 		nil, // Arguments
// 	)
// 	if err != nil {
// 		return amqp.Queue{}, err
// 	}

// 	return queue, nil

// }

// // BindQueue binds an existing queue to an existing exchange with a routing key
// func (rc *RabbitMQClient) CreateBinding(queueName, routingKey, exchangeName string) error {
// 	ch, err := rc.GetChannel()
// 	if err != nil {
// 		return err
// 	}
// 	defer ch.Close() // Close the channel after use

// 	return ch.QueueBind(
// 		queueName,    // Name of the queue to bind
// 		routingKey,   // Routing key for messages
// 		exchangeName, // Name of the exchange to bind to
// 		false,        // No wait
// 		nil,          // Arguments
// 	)
// }

// ! Consume

// func (rc *RabbitMQClient) Consume(queueName, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
// 	ch, err := rc.GetChannel()
// 	if err != nil {
// 		return nil, err
// 	}

// 	msgs, err := ch.Consume(
// 		queueName,
// 		consumer, // Consumer tag (can be left empty)
// 		autoAck,  // Auto-ack (set to false for manual ack)
// 		false,    // Exclusive (only this consumer can access the queue)
// 		false,    // No local (only deliver to this server)
// 		false,    // No wait
// 		nil,      // Arguments
// 	)
// 	if err != nil {
// 		// Consider closing the channel here if an error occurs during Consume
// 		ch.Close()
// 		return nil, err
// 	}

// 	// Close the channel in a separate deferred function to ensure it's closed
// 	// even if there are errors during message processing in the returned channel.
// 	defer ch.Close()

// 	return msgs, nil
// }

// !--?
// func (rc *RabbitMQClient) Consume(queueName, consumer string, autoAck bool) (<-chan amqp.Delivery, error) {
// 	ch, err := rc.GetChannel()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer ch.Close() // Close the channel after use

// 	return ch.Consume(
// 		queueName,
// 		consumer, // Consumer tag (can be left empty)
// 		autoAck,  // Auto-ack (set to false for manual ack)
// 		false,    // Exclusive (only this consumer can access the queue)
// 		false,    // No local (only deliver to this server)
// 		false,    // No wait
// 		nil,      // Arguments
// 	)
// }

// // ! PublishMessage sends a message to a specific exchange with a routing key
// func (rc *RabbitMQClient) PublishMessage(exchangeName string, routingKey string, options amqp.Publishing) error {
// 	ch, err := rc.GetChannel()
// 	if err != nil {
// 		return err
// 	}
// 	defer ch.Close() // Close the channel after use

// 	// body, err := json.Marshal(message) // Marshal the message to JSON
// 	// if err != nil {
// 	// 	return fmt.Errorf("failed to marshal message: %w", err)
// 	// }

// 	err = ch.Publish(
// 		exchangeName, // Name of the exchange
// 		routingKey,   // Routing key for message
// 		false,        // Mandatory (if true, message is rejected if no queue is bound)
// 		false,        // Immediate (if true, delivery happens now, or fails)
// 		options,
// 	)

// 	if err != nil {
// 		return err
// 	}

// 	//log.Println(confirmation.Wait())
// 	// confirmation.Wait()
// 	return nil

// }

//! QOS
// func (rc RabbitClient) ApplyQos(count, size int, global bool) error {
// 	return rc.ch.Qos(count, size, global)
// }
