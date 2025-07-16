package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
	"relayooor/api/pkg/logging"
)

type ActiveOperation struct {
	ID        string
	StartTime time.Time
	Type      string
}

type OperationTracker struct {
	mu         sync.RWMutex
	operations map[string]*ActiveOperation
}

func NewOperationTracker() *OperationTracker {
	return &OperationTracker{
		operations: make(map[string]*ActiveOperation),
	}
}

func (t *OperationTracker) Add(op *ActiveOperation) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.operations[op.ID] = op
}

func (t *OperationTracker) Remove(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.operations, id)
}

func (t *OperationTracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.operations)
}

func (t *OperationTracker) GetAll() []*ActiveOperation {
	t.mu.RLock()
	defer t.mu.RUnlock()
	
	ops := make([]*ActiveOperation, 0, len(t.operations))
	for _, op := range t.operations {
		ops = append(ops, op)
	}
	return ops
}

type GracefulServer struct {
	server          *http.Server
	shutdownTimeout time.Duration
	logger          *zap.Logger
	cleanup         []func() error
	tracker         *OperationTracker
}

func NewGracefulServer(server *http.Server, timeout time.Duration) *GracefulServer {
	return &GracefulServer{
		server:          server,
		shutdownTimeout: timeout,
		logger:          logging.With(zap.String("component", "server")),
		cleanup:         make([]func() error, 0),
		tracker:         NewOperationTracker(),
	}
}

func (s *GracefulServer) RegisterCleanup(fn func() error) {
	s.cleanup = append(s.cleanup, fn)
}

func (s *GracefulServer) GetTracker() *OperationTracker {
	return s.tracker
}

func (s *GracefulServer) ListenAndServe() error {
	// Channel to listen for interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Channel to track server errors
	serverErrors := make(chan error, 1)

	// Start server in goroutine
	go func() {
		s.logger.Info("Starting server", zap.String("addr", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Wait for interrupt signal or server error
	select {
	case err := <-serverErrors:
		return err
	case sig := <-stop:
		s.logger.Info("Received shutdown signal", zap.String("signal", sig.String()))
		return s.shutdown()
	}
}

func (s *GracefulServer) shutdown() error {
	s.logger.Info("Starting graceful shutdown",
		zap.Duration("timeout", s.shutdownTimeout))

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	// Wait for active clearing operations (max 5 minutes)
	if err := s.waitForActiveOperations(ctx); err != nil {
		s.logger.Warn("Some operations did not complete", zap.Error(err))
		// Persist state for recovery
		s.persistActiveOperations()
	}

	// Shutdown HTTP server
	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error("Failed to shutdown server gracefully", zap.Error(err))
		return err
	}

	// Run cleanup functions
	s.logger.Info("Running cleanup tasks", zap.Int("count", len(s.cleanup)))
	for i, cleanupFn := range s.cleanup {
		s.logger.Debug("Running cleanup task", zap.Int("index", i))
		if err := cleanupFn(); err != nil {
			s.logger.Error("Cleanup task failed",
				zap.Int("index", i),
				zap.Error(err))
		}
	}

	s.logger.Info("Graceful shutdown completed")
	return nil
}

func (s *GracefulServer) waitForActiveOperations(ctx context.Context) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	maxWait := 5 * time.Minute
	deadline := time.Now().Add(maxWait)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			count := s.tracker.Count()
			if count == 0 {
				return nil
			}

			if time.Now().After(deadline) {
				return fmt.Errorf("%d operations still active after %v", count, maxWait)
			}

			s.logger.Info("Waiting for active operations",
				zap.Int("count", count),
				zap.Duration("remaining", time.Until(deadline)))
		}
	}
}

func (s *GracefulServer) persistActiveOperations() {
	operations := s.tracker.GetAll()
	if len(operations) == 0 {
		return
	}

	s.logger.Info("Persisting active operations", zap.Int("count", len(operations)))
	
	// TODO: Implement persistence to database or file
	for _, op := range operations {
		s.logger.Info("Active operation",
			zap.String("id", op.ID),
			zap.String("type", op.Type),
			zap.Duration("age", time.Since(op.StartTime)))
	}
}