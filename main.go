package main

import (
	"context"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"entgo.io/ent/dialect/sql"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)

	secretDB := os.Getenv("SECRET")

	dbURI, err := getDBURI(secretDB)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dbURI)
	db, err := sqlx.Open("pgx", dbURI)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	driver := sql.OpenDB("postgres", db.DB)
	if err = driver.DB().Ping(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("ping to DB successfully")

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("defaulting to port %s", port)
	}

	// Start HTTP server.
	log.Printf("listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	name := os.Getenv("NAME")
	if name == "" {
		name = "mate"
	}
	fmt.Fprintf(w, "Hello %s!\n", name)
}

func getDBURI(name string) (string, error) {
	ctx := context.Background()
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create secretmanager client: %w", err)
	}
	defer client.Close()

	// Build the request.
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	// Call the API.
	result, err := client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %w", err)
	}

	// Verify the data checksum.
	crc32c := crc32.MakeTable(crc32.Castagnoli)
	checksum := int64(crc32.Checksum(result.Payload.Data, crc32c))
	if checksum != *result.Payload.DataCrc32C {
		return "", fmt.Errorf("Data corruption detected.")
	}

	return string(result.Payload.Data), nil
}
