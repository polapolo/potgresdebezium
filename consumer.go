package main

import (
	"context"
	"log"
	"time"

	"github.com/hamba/avro"
	"github.com/twmb/franz-go/pkg/kgo"
)

func InsertOrderConsumer() {
	db := connectDB(context.Background())
	defer db.Close()

	opts := []kgo.Opt{
		kgo.SeedBrokers(getRedPandaHosts()...),
		kgo.ConsumeTopics(
			topicOrdersInsertAVRO,
		),
		kgo.ConsumerGroup("insert_order_consumer"),
		kgo.GreedyAutoCommit(),
	}
	redPandaClient, err := kgo.NewClient(opts...)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer redPandaClient.Close()

	wholeStartTime := time.Now()
	schema := avro.MustParse(schemaOrder)

	var counter int64

consumerLoop:
	for {
		maxRecordPerConsume := 1000000
		fetches := redPandaClient.PollRecords(context.Background(), maxRecordPerConsume)
		for _, fetchErr := range fetches.Errors() {
			log.Printf("error consuming from topic: topic=%s, partition=%d, err=%v\n",
				fetchErr.Topic, fetchErr.Partition, fetchErr.Err)
			break consumerLoop
		}

		dataRows := make([][]interface{}, 0)
		records := fetches.Records()
		numOfRecords := len(records)

		for _, record := range fetches.Records() {
			counter++

			// parse avro to struct
			var order orderAVRO
			err := avro.Unmarshal(schema, record.Value, &order)
			if err != nil {
				log.Println(err)
				continue
			}

			// prepare data to be ingested
			row := []interface{}{
				order.ID,        // id
				order.UserID,    // user_id
				order.StockCode, // stock_code
				"B",             // type
				order.Lot,       // lot
				order.Price,     // price
				order.Status,    // status
				time.Now(),      // created_at
			}
			dataRows = append(dataRows, row)
		}

		// ingest / copy
		err = copyOrders(context.Background(), dataRows)
		if err != nil {
			log.Println(err)
			return
		}

		if numOfRecords > 0 {
			if counter == 1000000 {
				log.Println("Speed:", time.Since(wholeStartTime).Milliseconds(), "ms")
			}
		}
	}
}

func InsertOrderConsumerManualCommit() {
	db := connectDB(context.Background())
	defer db.Close()

	opts := []kgo.Opt{
		kgo.SeedBrokers(getRedPandaHosts()...),
		kgo.ConsumeTopics(
			topicOrdersInsertAVRO,
		),
		kgo.ConsumerGroup("insert_order_consumer"),
		kgo.AutoCommitMarks(),
	}
	redPandaClient, err := kgo.NewClient(opts...)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer redPandaClient.Close()

	wholeStartTime := time.Now()
	schema := avro.MustParse(schemaOrder)

	var counter int64

consumerLoop:
	for {
		maxRecordPerConsume := 2500
		fetches := redPandaClient.PollRecords(context.Background(), maxRecordPerConsume)
		for _, fetchErr := range fetches.Errors() {
			log.Printf("error consuming from topic: topic=%s, partition=%d, err=%v\n",
				fetchErr.Topic, fetchErr.Partition, fetchErr.Err)
			break consumerLoop
		}

		dataRows := make([][]interface{}, 0)
		records := fetches.Records()
		numOfRecords := len(records)

		// log.Println(numOfRecords)

		for _, record := range fetches.Records() {
			counter++

			// if counter == 1 {
			// 	log.Println("consumed")
			// }

			// parse avro to struct
			var order orderAVRO
			err := avro.Unmarshal(schema, record.Value, &order)
			if err != nil {
				log.Println(err)
				continue
			}

			// prepare data to be ingested
			row := []interface{}{
				order.ID,        // id
				order.UserID,    // user_id
				order.StockCode, // stock_code
				"B",             // type
				order.Lot,       // lot
				order.Price,     // price
				order.Status,    // status
				time.Now(),      // created_at
			}
			dataRows = append(dataRows, row)
		}

		// ingest / copy
		err = copyOrdersUnique(context.Background(), db, dataRows)
		if err != nil {
			log.Println(err)
			return
		}

		// mark records to be autocommitted
		redPandaClient.MarkCommitRecords(records...)

		if numOfRecords > 0 {
			if counter == 1000000 {
				log.Println("Speed:", time.Since(wholeStartTime).Milliseconds(), "ms")
			}
		}
	}
}

func InsertOrderConsumerSyncCommit() {
	db := connectDB(context.Background())
	defer db.Close()

	opts := []kgo.Opt{
		kgo.SeedBrokers(getRedPandaHosts()...),
		kgo.ConsumeTopics(
			topicOrdersInsertAVRO,
		),
		kgo.ConsumerGroup("insert_order_consumer"),
	}
	redPandaClient, err := kgo.NewClient(opts...)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	defer redPandaClient.Close()

	wholeStartTime := time.Now()
	schema := avro.MustParse(schemaOrder)

	var counter int64

consumerLoop:
	for {
		maxRecordPerConsume := 2500
		fetches := redPandaClient.PollRecords(context.Background(), maxRecordPerConsume)
		for _, fetchErr := range fetches.Errors() {
			log.Printf("error consuming from topic: topic=%s, partition=%d, err=%v\n",
				fetchErr.Topic, fetchErr.Partition, fetchErr.Err)
			break consumerLoop
		}

		dataRows := make([][]interface{}, 0)
		records := fetches.Records()
		numOfRecords := len(records)

		for _, record := range fetches.Records() {
			counter++

			// parse avro to struct
			var order orderAVRO
			err := avro.Unmarshal(schema, record.Value, &order)
			if err != nil {
				log.Println(err)
				continue
			}

			// prepare data to be ingested
			row := []interface{}{
				order.ID,        // id
				order.UserID,    // user_id
				order.StockCode, // stock_code
				"B",             // type
				order.Lot,       // lot
				order.Price,     // price
				order.Status,    // status
				time.Now(),      // created_at
			}
			dataRows = append(dataRows, row)
		}

		// ingest / copy
		err = copyOrdersUnique(context.Background(), db, dataRows)
		if err != nil {
			log.Println(err)
			return
		}

		// mark records to be autocommitted
		redPandaClient.MarkCommitRecords(records...)

		if numOfRecords > 0 {
			if counter == 1000000 {
				log.Println("Speed:", time.Since(wholeStartTime).Milliseconds(), "ms")
			}
		}
	}
}
