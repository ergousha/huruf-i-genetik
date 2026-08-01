package main
import (
"context"
"fmt"
"log"
"huruf-genetik/repository"
)
func main() {
ctx := context.Background()
repo, err := repository.NewNeo4jRepository("neo4j://localhost:7687", "neo4j", "password123")
if err != nil {
log.Fatalf("Failed to connect: %v", err)
}
defer repo.Close(ctx)
surahs, err := repo.GetAllSurahSignatures(ctx)
if err != nil {
log.Fatalf("Failed to get surahs: %v", err)
}
fmt.Printf("Total Surahs found in DB: %d\n", len(surahs))
}
