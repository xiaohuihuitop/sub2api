package handler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func newUsageRecordTestPool(t *testing.T) *service.UsageRecordWorkerPool {
	t.Helper()
	pool := service.NewUsageRecordWorkerPoolWithOptions(service.UsageRecordWorkerPoolOptions{
		WorkerCount:           1,
		QueueSize:             8,
		TaskTimeout:           time.Second,
		OverflowPolicy:        "drop",
		OverflowSamplePercent: 0,
		AutoScaleEnabled:      false,
	})
	t.Cleanup(pool.Stop)
	return pool
}

func TestGatewayHandlerSubmitUsageRecordTask_WithPool(t *testing.T) {
	pool := newUsageRecordTestPool(t)
	h := &GatewayHandler{usageRecordWorkerPool: pool}

	done := make(chan struct{})
	gotClientRequestID := make(chan any, 1)
	parent := context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-usage-1")
	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		gotClientRequestID <- ctx.Value(ctxkey.ClientRequestID)
		close(done)
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("task not executed")
	}
	require.Equal(t, "client-usage-1", <-gotClientRequestID)
}

func TestGatewayHandlerSubmitUsageRecordTask_DroppedPoolSyncFallback(t *testing.T) {
	pool := newUsageRecordTestPool(t)
	pool.Stop()
	h := &GatewayHandler{usageRecordWorkerPool: pool}
	var called atomic.Bool

	parent, cancel := context.WithCancel(context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-dropped-1"))
	cancel()

	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("expected deadline in fallback context")
		}
		require.NoError(t, ctx.Err())
		require.Equal(t, "client-dropped-1", ctx.Value(ctxkey.ClientRequestID))
		called.Store(true)
	})

	require.True(t, called.Load())
}

func TestGatewayHandlerSubmitUsageRecordTask_WithoutPoolSyncFallback(t *testing.T) {
	h := &GatewayHandler{}
	var called atomic.Bool

	parent, cancel := context.WithCancel(context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-fallback-1"))
	cancel()

	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("expected deadline in fallback context")
		}
		require.NoError(t, ctx.Err())
		require.Equal(t, "client-fallback-1", ctx.Value(ctxkey.ClientRequestID))
		called.Store(true)
	})

	require.True(t, called.Load())
}

func TestGatewayHandlerSubmitUsageRecordTask_NilTask(t *testing.T) {
	h := &GatewayHandler{}
	require.NotPanics(t, func() {
		h.submitUsageRecordTask(context.Background(), nil)
	})
}

func TestGatewayHandlerSubmitUsageRecordTask_WithoutPool_TaskPanicRecovered(t *testing.T) {
	h := &GatewayHandler{}
	var called atomic.Bool

	require.NotPanics(t, func() {
		h.submitUsageRecordTask(context.Background(), func(ctx context.Context) {
			panic("usage task panic")
		})
	})

	h.submitUsageRecordTask(context.Background(), func(ctx context.Context) {
		called.Store(true)
	})
	require.True(t, called.Load(), "panic 后后续任务应仍可执行")
}

func TestOpenAIGatewayHandlerSubmitUsageRecordTask_WithPool(t *testing.T) {
	pool := newUsageRecordTestPool(t)
	h := &OpenAIGatewayHandler{usageRecordWorkerPool: pool}

	done := make(chan struct{})
	gotClientRequestID := make(chan any, 1)
	parent := context.WithValue(context.Background(), ctxkey.ClientRequestID, "openai-client-usage-1")
	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		gotClientRequestID <- ctx.Value(ctxkey.ClientRequestID)
		close(done)
	})

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("task not executed")
	}
	require.Equal(t, "openai-client-usage-1", <-gotClientRequestID)
}

func TestOpenAIGatewayHandlerSubmitUsageRecordTask_DroppedPoolSyncFallback(t *testing.T) {
	pool := newUsageRecordTestPool(t)
	pool.Stop()
	h := &OpenAIGatewayHandler{usageRecordWorkerPool: pool}
	var called atomic.Bool

	parent, cancel := context.WithCancel(context.WithValue(context.Background(), ctxkey.ClientRequestID, "openai-client-dropped-1"))
	cancel()

	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("expected deadline in fallback context")
		}
		require.NoError(t, ctx.Err())
		require.Equal(t, "openai-client-dropped-1", ctx.Value(ctxkey.ClientRequestID))
		called.Store(true)
	})

	require.True(t, called.Load())
}

func TestOpenAIGatewayHandlerSubmitUsageRecordTask_WithoutPoolSyncFallback(t *testing.T) {
	h := &OpenAIGatewayHandler{}
	var called atomic.Bool

	parent, cancel := context.WithCancel(context.WithValue(context.Background(), ctxkey.ClientRequestID, "openai-client-fallback-1"))
	cancel()

	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("expected deadline in fallback context")
		}
		require.NoError(t, ctx.Err())
		require.Equal(t, "openai-client-fallback-1", ctx.Value(ctxkey.ClientRequestID))
		called.Store(true)
	})

	require.True(t, called.Load())
}

func TestOpenAIGatewayHandlerSubmitUsageRecordTask_NilTask(t *testing.T) {
	h := &OpenAIGatewayHandler{}
	require.NotPanics(t, func() {
		h.submitUsageRecordTask(context.Background(), nil)
	})
}

func TestOpenAIGatewayHandlerSubmitUsageRecordTask_WithoutPool_TaskPanicRecovered(t *testing.T) {
	h := &OpenAIGatewayHandler{}
	var called atomic.Bool

	require.NotPanics(t, func() {
		h.submitUsageRecordTask(context.Background(), func(ctx context.Context) {
			panic("usage task panic")
		})
	})

	h.submitUsageRecordTask(context.Background(), func(ctx context.Context) {
		called.Store(true)
	})
	require.True(t, called.Load(), "panic 后后续任务应仍可执行")
}
