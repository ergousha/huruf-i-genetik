package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"huruf-genetik/config"
	"huruf-genetik/repository"
	"huruf-genetik/services"
	"huruf-genetik/web"
)

func main() {
	// Let Docker Neo4j start
	log.Println("Waiting for Neo4j to be ready...")
	time.Sleep(5 * time.Second)

	// Neo4j Configuration
	neo4jURI := "bolt://localhost:7687"
	neo4jUser := "neo4j"
	neo4jPass := "password123"

	// Connect to Database with simple retry mechanism
	var db *config.Database
	var err error
	for i := 0; i < 5; i++ {
		db, err = config.NewDatabase(neo4jURI, neo4jUser, neo4jPass)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to Neo4j (attempt %d/5). Retrying in 3 seconds...", i+1)
		time.Sleep(3 * time.Second)
	}
	
	if err != nil {
		log.Fatalf("Could not connect to Neo4j after 5 attempts: %v", err)
	}
	defer db.Close(context.Background())

	// Initialize Repository
	repo := repository.NewNeo4jRepository(db)

	// Seed Database
	log.Println("Seeding Neo4j Database...")
	err = repo.Seed(context.Background())
	if err != nil {
		log.Printf("Error during seeding (might already be seeded): %v", err)
	} else {
		log.Println("Seeding successful.")
	}

	// Initialize Services
	analyzer := services.NewAnalyzerService()
	engine := services.NewEngineService(repo, analyzer)

	// Initialize Web Layer
	webHandler := web.NewWebHandler(engine)

	// Setup Routes
	http.HandleFunc("/", webHandler.ServeHome)
	http.HandleFunc("/analyze", webHandler.HandleAnalyze)
	http.HandleFunc("/api/dashboard", webHandler.HandleDashboard)
	http.HandleFunc("/api/import-quran", webHandler.HandleImportQuran)
	http.HandleFunc("/api/heatmap", webHandler.HandleAlemHeatmap)
	http.HandleFunc("/api/root-journey", webHandler.HandleRootJourney)
	http.HandleFunc("/api/esma-correlation", webHandler.HandleEsmaCorrelation)
	http.HandleFunc("/api/esma-matrix", webHandler.HandleEsmaMatrix)

	// Start HTTP Server
	port := "8080"
	log.Printf("Server starting on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
