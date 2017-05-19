package metrics

import (
	"math/rand"
	"os"
	"testing"
	"time"

	nats "github.com/nats-io/go-nats"
	"github.com/stretchr/testify/assert"
)

func generateMetric() M {
	return M{
		"uint64":  uint64(rand.Int()),
		"uint32":  uint32(rand.Int()),
		"unit16":  uint16(rand.Int()),
		"int64":   int64(rand.Int()),
		"int32":   int32(rand.Int()),
		"int":     rand.Int(),
		"float64": rand.Float64(),
		"float32": rand.Float32(),
	}
}

func TestMain(t *testing.M) {
	os.Setenv("APPLICATION_NAME", "default")
	t.Run()
}

func TestMetrics(t *testing.T) {
	t.Run("Unknown server", func(t *testing.T) {
		metrics, err := New("1.1.1.1:1111")

		assert.Error(t, err)
		assert.True(t, metrics == nil)
	})

	t.Run("Unknown server setup", func(t *testing.T) {
		err := Setup("1.1.1.1:1111")

		assert.Error(t, err)
		assert.True(t, DefaultConn == nil)
	})

	t.Run("Empty application", func(t *testing.T) {
		os.Setenv("APPLICATION_NAME", "")
		metrics, err := New("1.1.1.1:1111")

		assert.Error(t, err)
		assert.True(t, metrics == nil)
		os.Setenv("APPLICATION_NAME", "default")
	})

	t.Run("Empty queue", func(t *testing.T) {
		metrics, err := New("1.1.1.1:1111")

		assert.Error(t, err)
		assert.True(t, metrics == nil)
	})

	t.Run("Disabled metrics", func(t *testing.T) {
		metrics, err := New("")

		assert.NoError(t, err)
		assert.True(t, metrics != nil)

		err = metrics.SendAndWait(generateMetric())
		assert.NoError(t, err)
	})

	t.Run("Empty metrics", func(t *testing.T) {
		metrics, err := New(nats.DefaultURL)

		assert.NoError(t, err)
		assert.True(t, metrics != nil)

		err = metrics.SendAndWait(map[string]interface{}{})
		assert.NoError(t, err)
	})

	t.Run("Connection", func(t *testing.T) {
		metrics, err := New(nats.DefaultURL)

		assert.NoError(t, err)
		assert.True(t, metrics != nil)

		t.Run("Synchronous send", func(t *testing.T) {
			err := metrics.SendAndWait(generateMetric())
			assert.NoError(t, err)
		})

		t.Run("Asynchronous send", func(t *testing.T) {
			result := metrics.Send(generateMetric())
			err := <-result
			assert.NoError(t, err)
		})
	})

	t.Run("Default connection", func(t *testing.T) {
		err := Setup(nats.DefaultURL)

		assert.NoError(t, err)
		assert.True(t, DefaultConn != nil)

		t.Run("Synchronous send", func(t *testing.T) {
			err := SendAndWait(generateMetric())
			assert.NoError(t, err)
		})

		t.Run("Asynchronous send", func(t *testing.T) {
			result := Send(generateMetric())
			err := <-result
			assert.NoError(t, err)
		})
	})

	t.Run("Auto sending", func(t *testing.T) {
		metrics, err := New(nats.DefaultURL)

		assert.NoError(t, err)
		assert.True(t, metrics != nil)

		done := make(chan bool, 1)
		go func() {
			metrics.Watch(100)
			done <- true
		}()

		time.Sleep(time.Millisecond * 500)
		metrics.Disable()

		assert.True(t, <-done)
	})

	t.Run("Auto sending default connection", func(t *testing.T) {
		err := Setup(nats.DefaultURL)

		assert.NoError(t, err)
		assert.True(t, DefaultConn != nil)

		done := make(chan bool, 1)
		go func() {
			DefaultConn.Watch(100)
			done <- true
		}()

		time.Sleep(time.Millisecond * 500)
		DefaultConn.Disable()

		assert.True(t, <-done)
	})
}
