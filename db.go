package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

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

	return nil
}

func copyOrdersUnique(ctx context.Context, db *pgxpool.Pool, dataRows [][]interface{}) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		panic(err)
	}

	_, err = tx.Exec(ctx, "CREATE TEMP TABLE tmp_orders (LIKE orders INCLUDING DEFAULTS) ON COMMIT DROP;")
	if err != nil {
		panic(err)
	}

	cols := []string{"id", "user_id", "stock_code", "type", "lot", "price", "status", "created_at"}
	rows := pgx.CopyFromRows(dataRows)

	inserted, err := tx.CopyFrom(ctx, pgx.Identifier{"tmp_orders"}, cols, rows)
	if err != nil {
		panic(err)
	}

	if inserted != int64(len(dataRows)) {
		log.Printf("Failed to insert all the data! Expected: %d, Got: %d", len(dataRows), inserted)
		return errors.New("whut")
	}

	_, err = tx.Exec(ctx, `INSERT INTO orders SELECT * FROM tmp_orders ON CONFLICT DO NOTHING;`)
	if err != nil {
		panic(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		panic(err)
	}

	return nil
}
