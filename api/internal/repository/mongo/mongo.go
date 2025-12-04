package mongo

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client wraps the MongoDB client and database.
type Client struct {
	client *mongo.Client
	db     *mongo.Database
}

// NewClient creates a new MongoDB client and connects to the database.
func NewClient(uri, dbName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	// Verify connection with a ping
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	db := client.Database(dbName)

	log.Printf("Connected to MongoDB at %s, db: %s", uri, dbName)

	return &Client{
		client: client,
		db:     db,
	}, nil
}

// DB returns the underlying mongo.Database instance.
func (c *Client) DB() *mongo.Database {
	return c.db
}

// Collection returns a collection from the database.
func (c *Client) Collection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

// Close disconnects the MongoDB client.
func (c *Client) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

