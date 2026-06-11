// Demo application simulating a production Go microservice to test Goroviz.
// This spawns ~20 goroutines structured like a real-world app (DB pool, API
// handlers, cache eviction, Kafka consumers, and lock contention) to produce
// rich and interesting pprof profiles.
//
// Usage:
//   go run testdata/demo_app.go
//
// Then in another terminal, run Goroviz:
//   goroviz attach localhost:6060
//
package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"time"
)

func main() {
	fmt.Println("🚀 Starting simulated production microservice...")

	// 1. API Server (exposes pprof endpoint + simulates stuck/slow API calls)
	api := &HTTPServer{addr: "localhost:6060"}
	api.Start()

	// 2. Database Connection Pool (simulates connection starvation/waiting)
	db := &DatabasePool{}
	db.Start()

	// 3. Message Broker Consumer (simulates Kafka message consumption loops)
	broker := &KafkaConsumer{topic: "user-orders"}
	broker.Start()

	// 4. Redis Cache Sweeper (simulates periodic background tasks)
	cache := &RedisCache{}
	cache.Start()

	// 5. Shared Ledger / Order Book (simulates heavy mutex lock contention)
	ledger := &OrderProcessor{}
	ledger.Start()

	// 6. Metrics Exporter (simulates background scraping/exporting)
	metrics := &MetricsReporter{}
	metrics.Start()

	fmt.Println("\n📡 Services initialized:")
	fmt.Println("  • HTTP API Server        → Listening on http://localhost:6060")
	fmt.Println("  • pprof debug endpoint   → http://localhost:6060/debug/pprof/goroutine?debug=2")
	fmt.Println("  • Database Pool          → 5 queries waiting, 1 background opener")
	fmt.Println("  • Kafka Consumers        → 4 active consumer worker loops")
	fmt.Println("  • Cache Evictor          → 1 background cleaner loop")
	// Note: 1 goroutine holds the lock, 3 are blocked waiting.
	fmt.Println("  • Order Processor        → 3 threads blocked on Mutex lock contention")
	fmt.Println("  • Metrics Reporter       → 1 background publisher loop")
	fmt.Println("\nPress Ctrl+C to terminate.")

	// Keep service alive indefinitely
	select {}
}

// HTTPServer represents an API gateway.
type HTTPServer struct {
	addr string
}

func (s *HTTPServer) Start() {
	go func() {
		if err := http.ListenAndServe(s.addr, nil); err != nil {
			fmt.Printf("HTTP Server failed: %v\n", err)
		}
	}()

	// Simulate active, stuck API client calls (e.g. waiting for external payment gateway)
	for i := 0; i < 3; i++ {
		go s.handleRequest(i)
	}
}

//go:noinline
func (s *HTTPServer) handleRequest(reqID int) {
	s.processPayment(reqID)
}

//go:noinline
func (s *HTTPServer) processPayment(reqID int) {
	ch := make(chan struct{})
	<-ch // Blocked forever waiting for the fake payment gateway callback
}

// DatabasePool simulates a SQL database connection pool (e.g. pgx/sql.DB).
type DatabasePool struct {
	conns chan struct{}
}

func (db *DatabasePool) Start() {
	db.conns = make(chan struct{}, 10)

	// Spawn background connection opener
	go db.connectionOpener()

	// Spawn 5 transaction goroutines waiting to acquire a connection
	for i := 0; i < 5; i++ {
		go db.executeTransaction(i)
	}
}

//go:noinline
func (db *DatabasePool) connectionOpener() {
	for {
		time.Sleep(1 * time.Hour) // Waiting to open new connection requests
	}
}

//go:noinline
func (db *DatabasePool) executeTransaction(txID int) {
	db.acquireConnection(txID)
}

//go:noinline
func (db *DatabasePool) acquireConnection(txID int) {
	<-db.conns // Starved: waiting for a database connection to open up
}

// KafkaConsumer simulates message consumption loops.
type KafkaConsumer struct {
	topic string
}

func (kc *KafkaConsumer) Start() {
	// Spawn 4 parallel queue consumers
	for i := 0; i < 4; i++ {
		go kc.consumeLoop(i)
	}
}

//go:noinline
func (kc *KafkaConsumer) consumeLoop(workerID int) {
	messages := make(chan string)
	for {
		kc.processMessage(<-messages) // Blocked on channel read: waiting for incoming broker messages
	}
}

//go:noinline
func (kc *KafkaConsumer) processMessage(msg string) {
	_ = msg
}

// OrderProcessor simulates Mutex lock contention on a shared order book ledger.
type OrderProcessor struct {
	mu sync.Mutex
}

func (op *OrderProcessor) Start() {
	op.mu.Lock() // Acquire the lock forever in main to force contention

	// Spawn 3 workers that will block trying to lock the order book
	for i := 0; i < 3; i++ {
		go op.processOrder(i)
	}
}

//go:noinline
func (op *OrderProcessor) processOrder(orderID int) {
	op.lockOrderBook()
}

//go:noinline
func (op *OrderProcessor) lockOrderBook() {
	op.mu.Lock()
	defer op.mu.Unlock()
}

// RedisCache simulates a caching layer with background eviction.
type RedisCache struct{}

func (rc *RedisCache) Start() {
	go rc.evictionLoop()
}

//go:noinline
func (rc *RedisCache) evictionLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for range ticker.C {
		rc.clearExpiredKeys()
	}
}

//go:noinline
func (rc *RedisCache) clearExpiredKeys() {
	// Sweeper logic
}

// MetricsReporter simulates exporting metrics to Datadog/Prometheus.
type MetricsReporter struct{}

func (mr *MetricsReporter) Start() {
	go mr.reportLoop()
}

//go:noinline
func (mr *MetricsReporter) reportLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		mr.gatherAndPublish()
	}
}

//go:noinline
func (mr *MetricsReporter) gatherAndPublish() {
	// Publish logic
}
