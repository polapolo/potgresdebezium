FROM golang:1.17.8
WORKDIR /postgresbenchmark
COPY . /postgresbenchmark

# Install dependencies
RUN go mod vendor -v

# Compile
RUN go build -o main .

# Run
ENTRYPOINT "/postgresbenchmark/main"