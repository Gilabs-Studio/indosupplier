package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/gilabs/gims/api/internal/refresh_token/data/repositories"
)

// RefreshTokenCleanupWorker handles cleanup of expired refresh tokens
type RefreshTokenCleanupWorker struct {
	refreshTokenRepo repositories.RefreshTokenRepository
	ticker           *time.Ticker
	stopChan         chan struct{}
	stopOnce         sync.Once
}

// NewRefreshTokenCleanupWorker creates a new refresh token cleanup worker
func NewRefreshTokenCleanupWorker(
	refreshTokenRepo repositories.RefreshTokenRepository,
	interval time.Duration,
) *RefreshTokenCleanupWorker {
	return &RefreshTokenCleanupWorker{
		refreshTokenRepo: refreshTokenRepo,
		ticker:           time.NewTicker(interval),
		stopChan:         make(chan struct{}),
	}
}

// Start starts the refresh token cleanup worker
func (w *RefreshTokenCleanupWorker) Start() {
	log.Println("Refresh token cleanup worker started")

	go func() {
		for {
			select {
			case <-w.ticker.C:
				w.cleanupExpiredTokens()
			case <-w.stopChan:
				w.ticker.Stop()
				log.Println("Refresh token cleanup worker stopped")
				return
			}
		}
	}()
}

// Stop stops the refresh token cleanup worker
func (w *RefreshTokenCleanupWorker) Stop() {
	w.stopOnce.Do(func() {
		close(w.stopChan)
	})
}

// cleanupExpiredTokens deletes expired refresh tokens from the database
func (w *RefreshTokenCleanupWorker) cleanupExpiredTokens() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := w.refreshTokenRepo.DeleteExpired(ctx)
	if err != nil {
		log.Printf("Error cleaning up expired refresh tokens: %v", err)
		return
	}

	log.Println("Expired refresh tokens cleaned up successfully")
}
