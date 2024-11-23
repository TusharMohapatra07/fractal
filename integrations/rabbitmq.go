package integrations

import (
	"errors"
	"strings"

	"github.com/SkySingh04/fractal/interfaces"
	"github.com/SkySingh04/fractal/logger"
	"github.com/SkySingh04/fractal/registry"
	"github.com/streadway/amqp"
)

// RabbitMQSource struct represents the configuration for consuming messages from RabbitMQ.
type RabbitMQSource struct {
	URL       string `json:"rabbitmq_input_url"`
	QueueName string `json:"rabbitmq_input_queue_name"`
}

// RabbitMQDestination struct represents the configuration for publishing messages to RabbitMQ.
type RabbitMQDestination struct {
	URL       string `json:"rabbitmq_output_url"`
	QueueName string `json:"rabbitmq_output_queue_name"`
}

// FetchData connects to RabbitMQ, retrieves data, and passes it through validation and transformation pipelines.
func (r RabbitMQSource) FetchData(req interfaces.Request) (interface{}, error) {
	logger.Infof("Connecting to RabbitMQ Source: URL=%s, Queue=%s", req.RabbitMQInputURL, req.RabbitMQInputQueueName)

	if req.RabbitMQInputURL == "" || req.RabbitMQInputQueueName == "" {
		return nil, errors.New("missing RabbitMQ source details")
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(req.RabbitMQInputURL)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// Consume messages
	msgs, err := ch.Consume(
		req.RabbitMQInputQueueName, // queue
		"",                         // consumer
		true,                       // auto-ack
		false,                      // exclusive
		false,                      // no-local
		false,                      // no-wait
		nil,                        // args
	)
	if err != nil {
		return nil, err
	}

	// Process messages
	for msg := range msgs {
		logger.Infof("Message received from RabbitMQ: %s", msg.Body)

		// Validation
		validatedData, err := validateData(msg.Body)
		if err != nil {
			logger.Fatalf("Validation failed for message: %s, Error: %s", msg.Body, err)
			continue // Skip invalid message
		}

		// Transformation
		transformedData := transformData(validatedData)

		// Pass data to output
		err = routeToOutput(transformedData, req)
		if err != nil {
			logger.Fatalf("Error routing data to output: %s", err)
			continue
		}

		logger.Infof("Message successfully processed and routed: %s", transformedData)
	}

	return nil, errors.New("no messages processed")
}

// SendData connects to RabbitMQ and publishes data to the specified queue.
func (r RabbitMQDestination) SendData(data interface{}, req interfaces.Request) error {
	logger.Infof("Connecting to RabbitMQ Destination: URL=%s, Queue=%s", req.RabbitMQOutputURL, req.RabbitMQOutputQueueName)

	if req.RabbitMQOutputURL == "" || req.RabbitMQOutputQueueName == "" {
		return errors.New("missing RabbitMQ target details")
	}

	// Connect to RabbitMQ
	conn, err := amqp.Dial(req.RabbitMQOutputURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	// Declare the queue to ensure it exists
	_, err = ch.QueueDeclare(
		req.RabbitMQOutputQueueName, // queue name
		true,                        // durable
		false,                       // delete when unused
		false,                       // exclusive
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		return err
	}

	// Publish the message
	err = ch.Publish(
		"",                          // exchange
		req.RabbitMQOutputQueueName, // routing key
		false,                       // mandatory
		false,                       // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(data.(string)), // Assumes data is a string; modify as needed
		},
	)
	if err != nil {
		return err
	}

	logger.Infof("Message sent to RabbitMQ queue %s: %s", req.RabbitMQOutputQueueName, data)
	return nil
}

// Initialize the RabbitMQ integrations by registering them with the registry.
func init() {
	registry.RegisterSource("RabbitMQ", RabbitMQSource{})
	registry.RegisterDestination("RabbitMQ", RabbitMQDestination{})
}

// validateData ensures the input data meets the required criteria.
func validateData(data []byte) ([]byte, error) {
	logger.Infof("Validating data: %s", data)

	// Example: Check if data is non-empty
	if len(data) == 0 {
		return nil, errors.New("data is empty")
	}

	// Add custom validation logic here
	return data, nil
}

// transformData modifies the input data as per business logic.
func transformData(data []byte) []byte {
	logger.Infof("Transforming data: %s", data)

	// Example: Convert data to uppercase (modify as needed)
	transformed := []byte(strings.ToUpper(string(data)))
	return transformed
}

func routeToOutput(data []byte, req interfaces.Request) error {
	logger.Infof("Routing data to output: %s", data)

	outputHandler, exists := registry.GetDestination(req.Output)
	if !exists {
		return errors.New("invalid output destination")
	}

	if err := outputHandler.SendData(data, req); err != nil {
		return errors.New("failed to send data to output")
	}

	logger.Infof("Data successfully routed to %s", req.Output)
	return nil
}