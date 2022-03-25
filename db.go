package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	numOfPoolMaxConnection = "50"
)

func connectDB(ctx context.Context) *pgxpool.Pool {
	DATABASE_URL := "postgres://postgres:password@0.0.0.0:5432/benchmark?pool_max_conns=" + numOfPoolMaxConnection
	pgxConfig, err := pgxpool.ParseConfig(DATABASE_URL)
	if err != nil {
		log.Fatal(err)
	}

	dbPool, err := pgxpool.ConnectConfig(ctx, pgxConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	return dbPool
}

func copyOrders(ctx context.Context, dataRows [][]interface{}) error {
	db := connectDB(ctx)
	defer db.Close()

	startTime := time.Now()

	cols := []string{"id", "user_id", "stock_code", "type", "lot", "price", "status", "created_at"}

	rows := pgx.CopyFromRows(dataRows)
	inserted, err := db.CopyFrom(ctx, pgx.Identifier{"orders"}, cols, rows)
	if err != nil {
		panic(err)
	}

	if inserted != int64(len(dataRows)) {
		log.Printf("Failed to insert all the data! Expected: %d, Got: %d", len(dataRows), inserted)
		return errors.New("whut")
	}

	timeElapsed := time.Since(startTime)
	log.Println("Total Time Copy Order Speed:", timeElapsed.Milliseconds(), "ms")

	return nil
}
