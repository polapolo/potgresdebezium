package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/hamba/avro"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	ctx := context.Background()

	redPandaClient := getRedPandaClient()
	defer redPandaClient.Close()

	// consumer
	go InsertOrderConsumer()

	// http endpoint
	r := gin.Default()

	r.GET("/publish/orders/insert/avro", func(c *gin.Context) {
		numOfUserIDsString := c.DefaultQuery("numOfUserIDs", "1000")
		numOfUserIDs, _ := strconv.Atoi(numOfUserIDsString)
		numOfOrdersString := c.DefaultQuery("numOfOrders", "10")
		numOfOrders, _ := strconv.Atoi(numOfOrdersString)

		publishInsertOrderAVRO(redPandaClient, numOfUserIDs, numOfOrders)

		c.JSON(200, []interface{}{"DONE"})
	})

	r.GET("/consume/orders/avro", func(c *gin.Context) {
		var jsonString string

	consumerLoop:
		for {
			fetches := redPandaClient.PollFetches(ctx)
			iter := fetches.RecordIter()

			for _, fetchErr := range fetches.Errors() {
				log.Printf("error consuming from topic: topic=%s, partition=%d, err=%v\n",
					fetchErr.Topic, fetchErr.Partition, fetchErr.Err)
				break consumerLoop
			}

			schema := avro.MustParse(schemaOrder)

			for !iter.Done() {
				record := iter.Next()

				result := orderAVRO{}
				err := avro.Unmarshal(schema, record.Value, &result)
				if err != nil {
					log.Println(err)
				}
				resultJSON, _ := json.Marshal(result)
				jsonString = string(resultJSON)
				break consumerLoop
			}
		}

		c.JSON(200, []interface{}{jsonString})
	})

	r.GET("/publish/orders/upsert/avro", func(c *gin.Context) {
		numOfUserIDsString := c.DefaultQuery("numOfUserIDs", "1000")
		numOfUserIDs, _ := strconv.Atoi(numOfUserIDsString)
		numOfOrdersString := c.DefaultQuery("numOfOrders", "10")
		numOfOrders, _ := strconv.Atoi(numOfOrdersString)

		publishUpsertOrderAVRO(redPandaClient, numOfUserIDs, numOfOrders)

		c.JSON(200, []interface{}{"DONE"})
	})

	r.GET("/publish/trades/insert/avro", func(c *gin.Context) {
		numOfUserIDsString := c.DefaultQuery("numOfUserIDs", "1000")
		numOfUserIDs, _ := strconv.Atoi(numOfUserIDsString)
		numOfOrdersString := c.DefaultQuery("numOfOrders", "10")
		numOfOrders, _ := strconv.Atoi(numOfOrdersString)
		numOfTradesString := c.DefaultQuery("numOfTrades", "1")
		numOfTrades, _ := strconv.Atoi(numOfTradesString)

		publishInsertTradeAVRO(redPandaClient, numOfUserIDs, numOfOrders, numOfTrades)

		c.JSON(200, []interface{}{"DONE"})
	})

	r.GET("/consume/trades/avro", func(c *gin.Context) {
		var jsonString string

	consumerLoop:
		for {
			fetches := redPandaClient.PollFetches(ctx)
			iter := fetches.RecordIter()

			for _, fetchErr := range fetches.Errors() {
				log.Printf("error consuming from topic: topic=%s, partition=%d, err=%v\n",
					fetchErr.Topic, fetchErr.Partition, fetchErr.Err)
				break consumerLoop
			}

			schema := avro.MustParse(`{"type":"record","name":"trade","fields":[{"name":"order_id","type":"long"},{"name":"lot","type":"long"},{"name":"lot_multiplier","type":"int"},{"name":"price","type":"int"},{"name":"total","type":"long"},{"name":"created_at","type":"string"}]}`)

			for !iter.Done() {
				record := iter.Next()

				result := tradeAVRO{}
				err := avro.Unmarshal(schema, record.Value, &result)
				if err != nil {
					log.Println(err)
				}
				resultJSON, _ := json.Marshal(result)
				jsonString = string(resultJSON)
				break consumerLoop
			}
		}

		c.JSON(200, []interface{}{jsonString})
	})

	// example:
	// {"user_id":1,"stock_code":"BBCA","type":"B","lot":10,"price":1000,"status":1}
	// json: 77 bytes
	// avro: 12 bytes
	r.GET("/json_vs_avro", func(c *gin.Context) {
		myJSON, _ := json.Marshal(order{
			UserID:    1,
			StockCode: "BBCA",
			Type:      "B",
			Lot:       10,
			Price:     1000,
			Status:    1,
		})

		schema := avro.MustParse(`{"type":"record","name":"order","fields":[{"name":"user_id","type":"long"},{"name":"stock_code","type":"string"},{"name":"type","type":"string"},{"name":"lot","type":"long"},{"name":"price","type":"int"},{"name":"status","type":"int"}]}`)
		myAVRO, _ := avro.Marshal(schema,
			orderAVRO{
				UserID:    1,
				StockCode: "BBCA",
				Type:      "B",
				Lot:       10,
				Price:     1000,
				Status:    1,
			},
		)

		c.JSON(200, map[string]interface{}{
			"json": len(myJSON),
			"avro": len(myAVRO),
		})
	})

	r.Run(":8090")
}
