# Reference
- https://www.youtube.com/watch?v=YZRHqRznO-o&ab_channel=CodewithIrtiza
- https://github.com/irtiza07/postgres_debezium_cdc
- https://redpanda.com/blog/redpanda-debezium/

# Order Table
```
CREATE TABLE IF NOT EXISTS orders (
    id text,
    user_id BIGINT,
    stock_code varchar(6),
    type VARCHAR(1),
    lot BIGINT,
    price int,
    status int,
    created_at TIMESTAMP,
    PRIMARY KEY(id)
);
```

# Step By Step
1. Spin up all the services `sudo docker-compose up`
2. Create connector config `curl -i -X POST -H "Accept:application/json" -H "Content-Type:application/json" 127.0.0.1:8083/connectors/ --data "@connector/order-insert-connector.json"`
3. Create topic `orders_avro` using http://localhost:9000/topic/create
4. Run the service & consumer `make run`
5. Hit `http://localhost:8090/publish/orders/insert/avro?numOfUserIDs=100000&numOfOrders=10&numOfTrades=1` to publish 1 million order into topic `orders_avro` that need to be inserted.
6. The consumer will consume it, and use the `COPY` query to insert into db.
7. You can check the CDC here http://localhost:9000/topic/postgres.public.orders
