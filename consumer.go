package main

import (
	"context"
	"log"
	"time"

	"github.com/hamba/avro"
	"github.com/twmb/franz-go/pkg/kgo"
)

func InsertOrderConsumer() {
	opts := []kgo.Opt{
		kgo.SeedBrokers(getRedPandaHosts()...),
		kgo.ConsumeTopics(
			topicOrdersInsertAVRO,
		),
	}
	redPandaClient, err := kgo.NewClient(opts...)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	wholeStartTime := time.Now()
	schema := avro.MustParse(schemaOrder)

	var counter int64
	var loop int64

consumerLoop:
	for {
		loop++
		// startTime := time.Now()
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
		// var sumOfSpeed float64
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

			// if counter == 1000000 {
			// 	log.Printf("%d %+v", counter, order)
			// 	log.Println("Insert Order Speed:", time.Since(startTime).Nanoseconds())
			// }
		}

		// ingest / copy
		err = copyOrders(context.Background(), dataRows)
		if err != nil {
			log.Println(err)
			return
		}

		if numOfRecords > 0 {
			// speed := time.Since(startTime).Milliseconds()
			// sumOfSpeed += float64(speed)
			// log.Println("NumOfRecords:", numOfRecords, "; Speed:", speed, "ms")
			if counter == 1000000 {
				// log.Println("Avg Speed:", sumOfSpeed/float64(loop)/float64(maxRecordPerConsume), "ms")
				log.Println("Speed:", time.Since(wholeStartTime).Milliseconds(), "ms")
			}
		}
	}
}
