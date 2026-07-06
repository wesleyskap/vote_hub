package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/wesleyskap/orkai-runiq/v3/queue"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := "postgres://postgres:postgres@postgres:5432/bbb_development?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		fmt.Println("Ping error:", err)
		os.Exit(1)
	}

	storage, err := queue.NewPostgresStorage(db)
	if err != nil {
		panic(err)
	}

	env, err := storage.Dequeue(context.Background(), "votes_queue", nil)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	if env == nil {
		fmt.Println("No job found!")
	} else {
		fmt.Printf("Job found: %+v\n", env)
	}
}
