package reddit

import (
	"context"
	"time"
)

const defaultStreamInterval = time.Second * 5

type streamConfig[T Streamable] struct {
	Interval       time.Duration
	DiscardInitial bool
	MaxRequests    int

	UseDumbLogic  bool
	HighWaterMark HighWaterMark
	GetFunc       func(context.Context, string, string) ([]T, error)
}

func NewStreamConfig[T Streamable]() *streamConfig[T] {
	return &streamConfig[T]{
		Interval:       defaultStreamInterval,
		DiscardInitial: false,
		MaxRequests:    0,
		UseDumbLogic:   false,
		HighWaterMark:  NewHighWaterMark(10),
	}
}

// StreamOpt is a configuration option to configure a stream.
type StreamOpt[T Streamable] func(*streamConfig[T])

// WithStreamInterval sets the frequency at which data will be fetched for the stream.
// If the duration is 0 or less, it will not be set and the default will be used.
func WithStreamInterval[T Streamable](v time.Duration) StreamOpt[T] {
	return func(c *streamConfig[T]) {
		if v > 0 {
			c.Interval = v
		}
	}
}

// WithStreamDiscardInitial will discard data from the first fetch for the stream.
func WithStreamDiscardInitial[T Streamable]() StreamOpt[T] {
	return func(c *streamConfig[T]) {
		c.DiscardInitial = true
	}
}

// WithStreamMaxRequests sets a limit on the number of times data is fetched for a stream.
// If less than or equal to 0, it is assumed to be infinite.
func WithStreamMaxRequests[T Streamable](v int) StreamOpt[T] {
	return func(c *streamConfig[T]) {
		if v > 0 {
			c.MaxRequests = v
		}
	}
}

// WithStartFromFullID gives a basic HighWaterMark struct
func WithStartFromFullID[T Streamable](v string) StreamOpt[T] {
	return func(c *streamConfig[T]) {
		c.HighWaterMark = NewHighWaterMark(10, v)
	}
}

func WithHighWaterMark[T Streamable](capacity uint32, items ...string) StreamOpt[T] {
	return func(c *streamConfig[T]) {
		c.HighWaterMark = NewHighWaterMark(capacity, items...)
	}
}

func WithGetFunc[T Streamable](f func(context.Context, string, string) ([]T, error)) StreamOpt[T] {
	return func(c *streamConfig[T]) {
		c.GetFunc = f
	}
}

func WithDumbLogic[T Streamable]() StreamOpt[T] {
	return func(c *streamConfig[T]) {
		c.UseDumbLogic = true
	}
}
