package main

type order struct {
	ID        int64  `json:"id,omitempty" avro:"id,omitempty"`
	UserID    int64  `json:"user_id" avro:"user_id"`
	StockCode string `json:"stock_code" avro:"stock_code"`
	Type      string `json:"type" avro:"type"`
	Lot       int64  `json:"lot" avro:"lot"`
	Price     int    `json:"price" avro:"price"`
	Status    int    `json:"status" avro:"status"`
}

type orderAVRO struct {
	ID        string `db:"id" avro:"id"`
	UserID    int64  `db:"user_id" avro:"user_id"`
	StockCode string `db:"stock_code" avro:"stock_code"`
	Type      string `db:"type" avro:"type"`
	Lot       int64  `db:"lot" avro:"lot"`
	Price     int    `db:"price" avro:"price"`
	Status    int    `db:"status" avro:"status"`
}

type orderAVROUpsert struct {
	ID        string `avro:"id"`
	UserID    int64  `avro:"user_id"`
	StockCode string `avro:"stock_code"`
	Type      string `avro:"type"`
	Lot       int64  `avro:"lot"`
	Price     int    `avro:"price"`
	Status    int    `avro:"status"`
}

type tradeAVRO struct {
	OrderID       int64  `avro:"order_id"`
	Lot           int64  `avro:"lot"`
	LotMultiplier int    `avro:"lot_multiplier"`
	Price         int    `avro:"price"`
	Total         int64  `avro:"total"`
	CreatedAt     string `avro:"created_at"`
}
