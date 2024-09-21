package tasks

import (
	"context"
	"log"
	"time"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/auth"
)

// StartTokenCleanupTask initiates a background task to periodically
// clean up expired tokens from the blacklist.
// It runs a cleanup operation every hour using a ticker,
// ensuring that expired tokens are removed in a timely manner.
//
// The cleanup task is executed in a separate goroutine,
// and each cleanup run is given a timeout of 5 minutes to complete.
// If an error occurs during the cleanup process, it is logged for diagnostic purposes.
//
// Parameters:
//   - tb: a pointer to the TokenBlacklist object, used to perform the cleanup of expired tokens.
func StartTokenCleanupTask(tb *auth.TokenBlacklist) {
	// Create a new ticker that triggers every hour.
	// The ticker holds an internal channel (called C), which receives a "tick" (a signal) at regular intervals.
	// every hour, the ticker will send a message on its channel, which can be used to trigger some action.
	ticker := time.NewTicker(1 * time.Hour)

	// Start a new goroutine to run the token cleanup task in the background.
	go func() {
		// Continuously loop, waiting for signals from the ticker.
		for range ticker.C {
			// Create a context with a timeout of 5 minutes for each cleanup task.
			// This ensures that the cleanup process doesn't run indefinitely and stops if it exceeds 5 minutes.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

			// Call the CleanupExpired method to remove expired tokens from the blacklist using the context.
			// This method performs the actual cleanup in the database.
			err := tb.CleanupExpired(ctx)
			if err != nil {
				log.Printf("Error cleaning up expired tokens: %v", err)
			}

			// Call the cancel function to release resources associated with the context.
			// It's important to always cancel a context when it's no longer needed.
			cancel()
		}
	}()
}
