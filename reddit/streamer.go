package reddit

import (
	"context"
	"time"
)

const defaultStreamInterval = time.Second * 5

type streamConfig struct {
	Interval       time.Duration
	DiscardInitial bool
	MaxRequests    int

	StartFromFullID string
	GetFunc         func(context.Context, string, string) ([]Streamable, error)
}

// StreamOpt is a configuration option to configure a stream.
type StreamOpt func(*streamConfig)

// WithStreamInterval sets the frequency at which data will be fetched for the stream.
// If the duration is 0 or less, it will not be set and the default will be used.
func WithStreamInterval(v time.Duration) StreamOpt {
	return func(c *streamConfig) {
		if v > 0 {
			c.Interval = v
		}
	}
}

// WithStreamDiscardInitial will discard data from the first fetch for the stream.
func WithStreamDiscardInitial(c *streamConfig) {
	c.DiscardInitial = true
}

// WithStreamMaxRequests sets a limit on the number of times data is fetched for a stream.
// If less than or equal to 0, it is assumed to be infinite.
func WithStreamMaxRequests(v int) StreamOpt {
	return func(c *streamConfig) {
		if v > 0 {
			c.MaxRequests = v
		}
	}
}

func WithStartFromFullID(v string) StreamOpt {
	return func(c *streamConfig) {
		c.StartFromFullID = v
	}
}

func WithGetFunc(f func(context.Context, string, string) ([]Streamable, error)) StreamOpt {
	return func(c *streamConfig) {
		c.GetFunc = f
	}
}
