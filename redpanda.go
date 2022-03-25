package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/hamba/avro"
	"github.com/twmb/franz-go/pkg/kgo"
)

func getRedPandaHosts() []string {
	return []string{
		"0.0.0.0:9093",
	}
}

func getRedPandaClient() *kgo.Client {
	opts := []kgo.Opt{
		kgo.SeedBrokers(getRedPandaHosts()...),
		kgo.ConsumeTopics(
			topicOrdersInsertAVRO,
			topicOrdersUpsertAVRO,
			topicTradesInsertAVRO,
		),
	}
	redPandaClient, err := kgo.NewClient(opts...)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	return redPandaClient
}

const (
	topicOrdersInsertAVRO = "orders_avro"
	topicOrdersUpsertAVRO = "orders_upsert_avro"
	topicTradesInsertAVRO = "trades_avro"

	schemaOrder = `{"type":"record","name":"order","fields":[{"name":"id","type":"string"},{"name":"user_id","type":"long"},{"name":"stock_code","type":"string"},{"name":"type","type":"string"},{"name":"lot","type":"long"},{"name":"price","type":"int"},{"name":"status","type":"int"}]}`
	schemaTrade = `{"type":"record","name":"trade","fields":[{"name":"order_id","type":"long"},{"name":"lot","type":"long"},{"name":"lot_multiplier","type":"int"},{"name":"price","type":"int"},{"name":"total","type":"long"},{"name":"created_at","type":"string"}]}`
)

// insert orders avro
func generateInsertOrderAVRO(numOfUserIDs int, numOfOrders int) [][]byte {
	schema, err := avro.Parse(schemaOrder)
	if err != nil {
		log.Fatal(err)
	}

	counter := int64(0)

	orderAVROs := make([][]byte, 0)
	// users
	for i := 1; i <= numOfUserIDs; i++ {
		// orders
		for j := 1; j <= numOfOrders; j++ {
			orderType := "B"
			if j%2 == 0 {
				orderType = "S"
			}

			// offset := numOfOrders * (i - 1)

			counter++

			orderAVRO, err := avro.Marshal(
				schema,
				orderAVRO{
					// ID:        strconv.Itoa(j + offset),
					// ID:        strconv.FormatInt(time.Now().UnixMicro(), 10),
					// ID:        uuid.New().String(),
					ID:        strconv.FormatInt(counter, 10),
					UserID:    int64(i),
					StockCode: "BBCA",
					Type:      orderType,
					Lot:       10,
					Price:     1000,
					Status:    1,
				},
			)
			if err != nil {
				log.Println(err)
				continue
			}

			orderAVROs = append(orderAVROs, orderAVRO)
		}
	}

	return orderAVROs
}

func publishInsertOrderAVRO(redPandaClient *kgo.Client, numOfUserIDs int, numOfOrders int) {
	avros := generateInsertOrderAVRO(numOfUserIDs, numOfOrders)

	records := make([]*kgo.Record, 0)
	for _, myAVRO := range avros {
		records = append(records, &kgo.Record{Topic: topicOrdersInsertAVRO, Value: myAVRO})
	}

	err := redPandaClient.ProduceSync(context.Background(), records...).FirstErr()
	if err != nil {
		log.Println(err)
	}
}

// upsert orders avro
func generateUpsertOrderAVRO(numOfUserIDs int, numOfOrders int) [][]byte {
	schema, err := avro.Parse(schemaOrder)
	if err != nil {
		log.Fatal(err)
	}

	orderAVROs := make([][]byte, 0)
	// users
	for i := 1; i <= numOfUserIDs; i++ {
		// orders
		for j := 1; j <= numOfOrders; j++ {
			orderType := "B"
			if j%2 == 0 {
				orderType = "S"
			}

			offset := numOfOrders * (i - 1)

			orderAVRO, err := avro.Marshal(
				schema,
				orderAVROUpsert{
					ID:        strconv.Itoa(j + offset),
					UserID:    int64(i),
					StockCode: "BBCA",
					Type:      orderType,
					Lot:       10,
					Price:     1000,
					Status:    1,
				},
			)
			if err != nil {
				log.Println(err)
				continue
			}

			orderAVROs = append(orderAVROs, orderAVRO)
		}
	}

	return orderAVROs
}

func publishUpsertOrderAVRO(redPandaClient *kgo.Client, numOfUserIDs int, numOfOrders int) {
	avros := generateUpsertOrderAVRO(numOfUserIDs, numOfOrders)

	records := make([]*kgo.Record, 0)
	for _, myAVRO := range avros {
		records = append(records, &kgo.Record{Topic: topicOrdersUpsertAVRO, Value: myAVRO})
	}

	err := redPandaClient.ProduceSync(context.Background(), records...).FirstErr()
	if err != nil {
		log.Println(err)
	}
}

// insert trades avro
func generateInsertTradeAVRO(numOfUserIDs int, numOfOrders int, numOfTrades int) [][]byte {
	schema, err := avro.Parse(schemaTrade)
	if err != nil {
		log.Fatal(err)
	}

	tradeAVROs := make([][]byte, 0)
	// users
	for i := 1; i <= numOfUserIDs; i++ {
		// orders
		for j := 1; j <= numOfOrders; j++ {
			offset := numOfOrders * (i - 1)

			// trades
			for k := 1; k <= numOfTrades; k++ {
				tradeAVRO, err := avro.Marshal(
					schema,
					tradeAVRO{
						OrderID:       int64(j + offset),
						Lot:           10,
						LotMultiplier: 100,
						Price:         1000,
						Total:         1000000,
						CreatedAt:     time.Now().Format(time.RFC3339),
					},
				)
				if err != nil {
					log.Println(err)
					continue
				}

				tradeAVROs = append(tradeAVROs, tradeAVRO)
			}
		}
	}

	return tradeAVROs
}

func publishInsertTradeAVRO(redPandaClient *kgo.Client, numOfUserIDs int, numOfOrders int, numOfTrades int) {
	avros := generateInsertTradeAVRO(numOfUserIDs, numOfOrders, numOfTrades)

	records := make([]*kgo.Record, 0)
	for _, myAVRO := range avros {
		records = append(records, &kgo.Record{Topic: topicTradesInsertAVRO, Value: myAVRO})
	}

	err := redPandaClient.ProduceSync(context.Background(), records...).FirstErr()
	if err != nil {
		log.Println(err)
	}
}
