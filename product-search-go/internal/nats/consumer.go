package nats

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/example/product-search/internal/service"
	"github.com/nats-io/nats.go"
)

type Consumer struct {
	natsURL string
	service *service.ProductSearchService
	conn    *nats.Conn
	sub     *nats.Subscription
	wg      sync.WaitGroup
}

func NewConsumer(natsURL string, service *service.ProductSearchService) *Consumer {
	return &Consumer{
		natsURL: natsURL,
		service: service,
	}
}

func (c *Consumer) Start() error {
	var err error

	// Connect to NATS with retry
	for i := 0; i < 10; i++ {
		c.conn, err = nats.Connect(c.natsURL,
			nats.Timeout(5*time.Second),
			nats.ReconnectWait(2*time.Second),
			nats.MaxReconnects(10),
		)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to NATS (attempt %d): %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return err
	}

	log.Printf("Connected to NATS at %s", c.natsURL)

	// Subscribe to Debezium CDC events
	// Subject format from Debezium: product.public.products
	subject := "product.public.products"

	c.sub, err = c.conn.Subscribe(subject, c.handleMessage)
	if err != nil {
		return err
	}

	log.Printf("NATS consumer started, listening on: %s", subject)

	// Keep the consumer running
	c.wg.Add(1)
	select {}
}

func (c *Consumer) handleMessage(msg *nats.Msg) {
	ctx := context.Background()

	log.Printf("Received CDC event: %s", string(msg.Data))

	// Parse Debezium JSON envelope
	var envelope struct {
		Payload struct {
			Op     string                 `json:"op"`
			Before map[string]interface{} `json:"before"`
			After  map[string]interface{} `json:"after"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(msg.Data, &envelope); err != nil {
		log.Printf("Failed to parse CDC event: %v", err)
		msg.Nak()
		return
	}

	operation := envelope.Payload.Op
	after := envelope.Payload.After
	before := envelope.Payload.Before

	switch operation {
	case "c", "r": // Create, Read (snapshot)
		if productID, ok := after["id"].(string); ok {
			log.Printf("Indexing new product: %s", productID)
			if err := c.service.IndexProduct(ctx, productID); err != nil {
				log.Printf("Failed to index product: %v", err)
				msg.Nak()
				return
			}
		}
	case "u": // Update
		if productID, ok := after["id"].(string); ok {
			log.Printf("Updating product index: %s", productID)
			if err := c.service.IndexProduct(ctx, productID); err != nil {
				log.Printf("Failed to update product index: %v", err)
				msg.Nak()
				return
			}
		}
	case "d": // Delete
		if before != nil {
			if productID, ok := before["id"].(string); ok {
				log.Printf("Removing product from index: %s", productID)
				if err := c.service.RemoveProduct(ctx, productID); err != nil {
					log.Printf("Failed to remove product: %v", err)
					msg.Nak()
					return
				}
			}
		}
	default:
		log.Printf("Unknown operation: %s", operation)
	}

	msg.Ack()
}

func (c *Consumer) Stop() {
	if c.sub != nil {
		c.sub.Unsubscribe()
	}
	if c.conn != nil {
		c.conn.Close()
	}
	c.wg.Done()
	log.Println("NATS consumer stopped")
}
