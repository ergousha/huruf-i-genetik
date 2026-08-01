package config

import (
	"context"
	"fmt"
	"log"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// Database holds the Neo4j driver connection
type Database struct {
	Driver neo4j.DriverWithContext
}

// NewDatabase initializes a new connection to Neo4j
func NewDatabase(uri, username, password string) (*Database, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, fmt.Errorf("failed to create Neo4j driver: %w", err)
	}

	// Verify connectivity
	ctx := context.Background()
	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Neo4j: %w", err)
	}

	log.Println("Successfully connected to Neo4j!")

	return &Database{Driver: driver}, nil
}

// Close gracefully closes the Neo4j driver connection
func (db *Database) Close(ctx context.Context) error {
	return db.Driver.Close(ctx)
}
