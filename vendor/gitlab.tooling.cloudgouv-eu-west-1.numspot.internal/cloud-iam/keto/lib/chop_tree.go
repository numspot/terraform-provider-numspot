package lib

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

const DefaultBufferSize uint64 = 500

// ChopClient will truncate a keto graph from a specific node.
type ChopClient struct {
	readService  Reader
	writeService Writer
}

type Chopper interface {
	Chop(ctx context.Context, namespace Namespace, object string, buffSize *uint64) error
}

// NewChopClient returns a ready to use [ChopClient].
func NewChopClient(readClient Reader, writeClient Writer) Chopper {
	return &ChopClient{
		readService:  readClient,
		writeService: writeClient,
	}
}

// Chop will truncate a keto relation tree entirely from the provided starting point, starting from the leaves.
// Bear in mind the context deadline needs to be long enough as to not interrupt the operation.
// This call may take a bit to run.
func (client *ChopClient) Chop(ctx context.Context, namespace Namespace, object string, buffSize *uint64) error {
	var bufferLength = DefaultBufferSize
	if buffSize != nil {
		bufferLength = *buffSize
	}

	tuple := &RelationTuple{
		Namespace: namespace,
		Object:    object,
	}

	tupleChan := make(chan *RelationTuple, int(bufferLength))
	finishedChan := make(chan bool)

	go client.work(ctx, tupleChan, finishedChan, bufferLength)
	m := new(sync.Map)
	defer func(tupleChan chan<- *RelationTuple, finished <-chan bool) {
		close(tupleChan)
		<-finishedChan
	}(tupleChan, finishedChan)

	if err := client.chop(ctx, tuple, tupleChan, m, 0); err != nil {
		return fmt.Errorf("client.chop: %w", err)
	}

	return nil
}

// chop is a recursive function that does the heavy lifting for [ChopClient.Chop].
func (client *ChopClient) chop(ctx context.Context, tuple *RelationTuple, c chan<- *RelationTuple, m *sync.Map, depth int) error {
	var cursor *string
	query := &RelationQuery{
		Subject: &Subject{
			Set: &SubjectSet{
				Namespace: tuple.Namespace,
				Object:    tuple.Object,
			},
		},
	}

	fmt.Sprintln(*query.Subject.Set)

	if depth == 0 {
		err := client.writeService.DeleteRelationTuplesFromQuery(ctx, RelationQuery{
			Namespace: &tuple.Namespace,
			Object:    &tuple.Object,
		})
		if err != nil {
			return fmt.Errorf("client.writeService.DeleteRelationTuplesFromQuery: %w", err)
		}
	}

	for {
		tuples, next, err := client.readService.ListRelationTuples(ctx, query, nil, cursor)
		if err != nil {
			return fmt.Errorf("client.readService.ListRelationTuples: %w", err)
		}

		cursor = &next

		skipped := 0
		for i := range tuples {
			_, loaded := m.LoadOrStore(tuples[i].String(), 0)
			if loaded {
				skipped++
			}
		}

		if len(tuples) == 0 || skipped == len(tuples) {
			if depth != 0 {
				c <- tuple
			}

			return nil
		}

		for i := range tuples {
			if err := client.chop(ctx, &tuples[i], c, m, depth+1); err != nil {
				return fmt.Errorf("client.chop: %w", err)
			}
		}
	}
}

// work will consume all values in a channel and once it has read buffSize values delete them.
// it will also delete the remaining values after channel is closed.
// remember to close the channel, lest the goroutine be leaked
func (client *ChopClient) work(ctx context.Context, c <-chan *RelationTuple, finished chan<- bool, buffSize uint64) {
	defer func(finished chan<- bool) {
		finished <- true
	}(finished)
	attrs := []slog.Attr{
		slog.Uint64("buffer_size", buffSize),
		slog.String("func", "ChopClient.work"),
	}

	buffer := make([]RelationTuple, 0, buffSize)

	start := time.Now()
	var count uint64

	for tuple := range c {
		buffer = append(buffer, *tuple)
		if len(buffer) != int(buffSize) {
			continue
		}

		if err := client.writeService.DeleteRelationTuples(ctx, buffer); err != nil {
			slog.LogAttrs(ctx, slog.LevelError, "error deleting relations tuples", append(attrs, slog.Any("tuples", buffer), slog.String("err", err.Error()))...)
		}

		count += buffSize

		buffer = buffer[:0]
	}

	count += uint64(len(buffer))
	if len(buffer) == 0 {
		attrs = append(attrs, slog.Duration("truncation_duration", time.Since(start)), slog.Uint64("count", count))

		slog.LogAttrs(ctx, slog.LevelInfo, "truncated graph", attrs...)

		return
	}

	if err := client.writeService.DeleteRelationTuples(ctx, buffer); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "error deleting relations tuples", append(attrs, slog.Any("tuples", buffer))...)
	}

	attrs = append(attrs, slog.Duration("truncation_duration", time.Since(start)), slog.Uint64("count", count))

	slog.LogAttrs(ctx, slog.LevelInfo, "truncated graph", attrs...)
}
