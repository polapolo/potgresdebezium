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

	startTime := time.Now()
	schema := avro.MustParse(schemaOrder)

	var counter int64

consumerLoop:
	for {
		fetches := redPandaClient.PollRecords(context.Background(), 1000000)
		for _, fetchErr := range fetches.Errors() {
			log.Printf("error consuming from topic: topic=%s, partition=%d, err=%v\n",
				fetchErr.Topic, fetchErr.Partition, fetchErr.Err)
			break consumerLoop
		}

		dataRows := make([][]interface{}, 0)
		records := fetches.Records()
		log.Println("Num. of Records:", len(records))
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
		log.Println("Insert Order Speed:", time.Since(startTime).Milliseconds())
	}
}
