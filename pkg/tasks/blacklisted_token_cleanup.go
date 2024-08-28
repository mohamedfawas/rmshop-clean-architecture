package tasks

import (
	"context"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
)

func StartTokenCleanupTask(tb *auth.TokenBlacklist) {
	ticker := time.NewTicker(1 * time.Hour)
	go func() {
		for range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			err := tb.CleanupExpired(ctx)
			if err != nil {
				log.Printf("Error cleaning up expired tokens: %v", err)
			}
			cancel()
		}
	}()
}
