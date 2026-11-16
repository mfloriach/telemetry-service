package relay

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"

	"server/internal/ingest"
)

// Publisher publishes a telemetry record value to Kafka.
type Publisher interface {
	Publish(ctx context.Context, value []byte) error
	Close() error
}

// KafkaPublisher publishes messages to a Kafka topic using segmentio/kafka-go.
type KafkaPublisher struct {
	writer *kafka.Writer
}

// NewKafkaPublisher creates a Kafka publisher for the given brokers and topic.
func NewKafkaPublisher(brokers []string, topic string) (*KafkaPublisher, error) {
	if len(brokers) == 0 {
		return nil, errors.New("at least one Kafka broker is required")
	}
	if strings.TrimSpace(topic) == "" {
		return nil, errors.New("Kafka topic is required")
	}
	w := &kafka.Writer{
		Addr:     kafka.TCP(brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaPublisher{writer: w}, nil
}

// Publish writes a message with the provided value to the configured topic.
func (p *KafkaPublisher) Publish(ctx context.Context, value []byte) error {
	if p == nil || p.writer == nil {
		return errors.New("kafka publisher is not initialized")
	}
	return p.writer.WriteMessages(ctx, kafka.Message{Value: value})
}

// Close closes the underlying Kafka writer.
func (p *KafkaPublisher) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

// Runner consumes telemetry records from an upstream WebSocket and publishes them to Kafka.
type Runner struct {
	dialer     ingest.Dialer
	publisher  Publisher
	logger     *slog.Logger
	baseDelay  time.Duration
	maxDelay   time.Duration
	successMin time.Duration
	jitter     func(time.Duration) time.Duration
	sleep      func(context.Context, time.Duration) error
}

// NewRunner creates a relay runner.
func NewRunner(dialer ingest.Dialer, publisher Publisher, logger *slog.Logger) *Runner {
	if logger == nil {
		logger = slog.Default()
	}
	return &Runner{
		dialer:     dialer,
		publisher:  publisher,
		logger:     logger,
		baseDelay:  100 * time.Millisecond,
		maxDelay:   5 * time.Second,
		successMin: time.Second,
		jitter: func(d time.Duration) time.Duration {
			if d <= 0 {
				return 0
			}
			half := d / 2
			if half <= 0 {
				return d
			}
			return half + rand.N(d-half+1) //nolint:gosec // non-cryptographic jitter for reconnect backoff
		},
		sleep: sleepWithContext,
	}
}

// Run maintains the relay loop until ctx is canceled.
func (r *Runner) Run(ctx context.Context, wsURL string) error {
	if r.dialer == nil {
		return errors.New("websocket dialer is required")
	}
	if r.publisher == nil {
		return errors.New("publisher is required")
	}

	attempt := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		connectedAt := time.Now()
		err := r.runOnce(ctx, wsURL)
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return ctx.Err()
		}
		if err != nil {
			r.logger.WarnContext(ctx, "relay upstream disconnected", "url", wsURL, "error", err)
			if time.Since(connectedAt) >= r.successMin {
				attempt = 0
			}
			delay := r.backoffDelay(attempt)
			attempt++
			r.logger.InfoContext(ctx, "relay reconnect scheduled", "url", wsURL, "attempt", attempt, "delay", delay.String())
			if sleepErr := r.sleep(ctx, delay); sleepErr != nil {
				return ctx.Err()
			}
			continue
		}
	}
}

func (r *Runner) runOnce(ctx context.Context, wsURL string) error {
	conn, err := r.dialer.Dial(ctx, wsURL)
	if err != nil {
		return fmt.Errorf("dial upstream websocket: %w", err)
	}
	defer func() {
		_ = conn.Close(1000, "relay stopping")
	}()

	r.logger.InfoContext(ctx, "relay upstream connected", "url", wsURL)
	for {
		msg, err := conn.Read(ctx)
		if err != nil {
			return err
		}
		if err := r.publisher.Publish(ctx, msg.Data); err != nil {
			return fmt.Errorf("publish telemetry record to kafka: %w", err)
		}
	}
}

func (r *Runner) backoffDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	exp := math.Pow(2, float64(attempt))
	d := time.Duration(float64(r.baseDelay) * exp)
	if d <= 0 || d > r.maxDelay {
		d = r.maxDelay
	}
	if r.jitter != nil {
		return r.jitter(d)
	}
	return d
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
