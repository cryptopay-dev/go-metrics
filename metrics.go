package metrics

import (
	"encoding/json"
	"errors"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/nats-io/go-nats"
)

type conn struct {
	mu          sync.RWMutex
	nats        *nats.Conn
	enabled     bool
	queue       string
	url         string
	hostname    string
	application string
}

// M metrics storage
// Example:
// m := metrics.M{
// 	"metric": 1000,
//	"gauge": 1,
//	"tag": "some_default_tag"
// }
type M map[string]interface{}

// DefaultConn shared default metric
// connection
var DefaultConn *conn

// Setup rewrites default metrics configuration
//
// Params:
// - url (in e.g. "nats://localhost:4222")
// - queue (in e.g. "metrics")
// - options nats.Option array
//
// Example:
// import (
//     "log"
//
//     "github.com/cryptopay.dev/go-metrics"
// )
//
// func main() {
//     err := metrics.Setup("nats://localhost:4222", "metrics")
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     for i:=0; i<10; i++ {
//         err = metrics.SendAndWait(metrics.M{
//             "counter": i,
//         })
//
//         if err != nil {
//             log.Fatal(err)
//         }
//     }
// }
func Setup(url, queue string, options ...nats.Option) error {
	metrics, err := New(url, queue, options...)
	if err != nil {
		return err
	}

	DefaultConn = metrics
	return nil
}

// New creates new metrics connection
//
// Params:
// - url (in e.g. "nats://localhost:4222")
// - queue (in e.g. "metrics")
// - options nats.Option array
//
// Example:
// import (
//     "log"
//
//     "github.com/cryptopay.dev/go-metrics"
// )
//
// func main() {
//     m, err := metrics.New("nats://localhost:4222", "metrics")
//     if err != nil {
//         log.Fatal(err)
//     }
//
//     for i:=0; i<10; i++ {
//         err = m.SendAndWait(metrics.M{
//             "counter": i,
//         })
//
//         if err != nil {
//             log.Fatal(err)
//         }
//     }
// }
func New(url, queue string, options ...nats.Option) (*conn, error) {
	if os.Getenv("INFLUX_METRICS_ENABLED") == "" {
		return &conn{
			enabled: false,
		}, nil
	}

	// Getting current environment
	app := os.Getenv("APPLICATION_NAME")
	if app == "" {
		return nil, errors.New("Application name not set")
	}

	if queue == "" {
		return nil, errors.New("Queue cannot be empty")
	}

	// Getting hostname up
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	nc, err := nats.Connect(url, options...)
	if err != nil {
		return nil, err
	}

	conn := &conn{
		nats:        nc,
		hostname:    hostname,
		application: app,
		enabled:     true,
		queue:       queue,
	}

	return conn, nil
}

// Send metrics to NATS queue
//
// Example:
// m.Send(metrics.M{
// 		"counter": i,
// })
func Send(metrics M) (err chan error) {
	return DefaultConn.Send(metrics)
}

// SendAndWait metrics to NATS queue waiting for response
//
// Example:
// err = m.SendAndWait(metrics.M{
// 		"counter": i,
// })
func SendAndWait(metrics M) error {
	return DefaultConn.SendAndWait(metrics)
}

// Send metrics to NATS queue
//
// Example:
// m.Send(metrics.M{
// 		"counter": i,
// })
func (m *conn) Send(metrics M) chan error {
	ch := make(chan error, 1)

	go func() {
		ch <- m.SendAndWait(metrics)
	}()

	return ch
}

// SendAndWait metrics to NATS queue waiting for response
//
// Example:
// err = m.SendAndWait(metrics.M{
// 		"counter": i,
// })
func (m *conn) SendAndWait(metrics M) error {
	m.mu.RLock()
	if !m.enabled {
		m.mu.RUnlock()
		return nil
	}
	m.mu.RUnlock()

	if len(metrics) == 0 {
		return nil
	}

	m.mu.RLock()
	metrics["hostname"] = m.hostname
	metrics["app"] = m.application
	m.mu.RUnlock()

	buf, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	m.mu.RLock()
	queue := m.queue
	m.mu.RUnlock()

	return m.nats.Publish(queue, buf)
}

// Disable disables watcher and disconnects
func (m *conn) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.enabled = false
	m.nats.Close()
}

// Watch watches memory, goroutine counter
func (m *conn) Watch(interval int) {
	var mem runtime.MemStats

	for {
		m.mu.RLock()
		enabled := m.enabled
		m.mu.RUnlock()

		if !enabled {
			break
		}

		// Getting memory stats
		runtime.ReadMemStats(&mem)

		metric := M{
			"alloc":       mem.Alloc,
			"heap_alloc":  mem.HeapAlloc,
			"total_alloc": mem.TotalAlloc,
			"gorotines":   runtime.NumGoroutine(),
		}
		m.SendAndWait(metric)

		time.Sleep(time.Millisecond * time.Duration(interval))
	}
}