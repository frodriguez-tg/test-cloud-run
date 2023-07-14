package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"entgo.io/ent/dialect/sql"

	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	log.Print("starting server...")
	http.HandleFunc("/", handler)

	dbURI := os.Getenv("V3_DBURL")

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
