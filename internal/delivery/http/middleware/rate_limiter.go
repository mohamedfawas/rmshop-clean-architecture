package middleware

import (
	"net/http"
	"sync"

	"github.com/mohamedfawas/rmshop-clean-architecture/pkg/api"
	"golang.org/x/time/rate"
)

// IPRateLimiter is a struct that handles rate limiting based on IP addresses.
// It maintains a map of IP addresses to their rate limiters, a mutex for concurrent access,
// and rate limit configurations.
type IPRateLimiter struct {
	ips map[string]*rate.Limiter // A map to hold the rate limiters for each IP address
	mu  *sync.RWMutex            // A mutex to allow safe concurrent access to the map
	r   rate.Limit               // The rate limit (requests per second)
	b   int                      // The burst size (maximum number of requests allowed at once)
}

// NewIPRateLimiter initializes and returns an IPRateLimiter.
// It takes the rate limit and burst size as arguments.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips: make(map[string]*rate.Limiter), // Initialize the map to hold IPs and their rate limiters
		mu:  &sync.RWMutex{},                // Initialize the mutex for concurrent access
		r:   r,                              // Set the rate limit
		b:   b,                              // Set the burst size
	}

	return i // Return the new instance of IPRateLimiter
}

// AddIP creates a new rate limiter for the given IP address and adds it to the map.
// It acquires a write lock on the mutex to ensure safe access.
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()         // Lock the mutex to prevent concurrent writes
	defer i.mu.Unlock() // Unlock the mutex once the IP limiter is added

	limiter := rate.NewLimiter(i.r, i.b) // Create a new rate limiter for the IP using the provided rate limit and burst size

	i.ips[ip] = limiter // Store the new rate limiter in the map, associated with the given IP address

	return limiter
}

// GetLimiter retrieves the rate limiter for the given IP address.
// If the rate limiter doesn't exist for the IP, it creates one.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()                  // Lock the mutex to safely access the map
	limiter, exists := i.ips[ip] // Check if the rate limiter already exists for the IP

	// If the rate limiter doesn't exist, create a new one for the IP
	if !exists {
		i.mu.Unlock()      // Unlock the mutex before calling AddIP (to avoid deadlocks)
		return i.AddIP(ip) // Create and return a new limiter for the IP
	}

	i.mu.Unlock() // Unlock the mutex after reading the map

	return limiter // Return the existing rate limiter for the IP
}

// RateLimitMiddleware is an HTTP middleware function that applies rate limiting.
// It checks the request's IP address and ensures that it doesn't exceed the rate limit.
// If the rate limit is exceeded, it returns an HTTP 429 Too Many Requests response.
func RateLimitMiddleware(next http.HandlerFunc, limiter *IPRateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr // Extract the client's IP address from the request

		// Check if the request from this IP is allowed by the rate limiter
		if !limiter.GetLimiter(ip).Allow() {
			// If not allowed, return a 429 Too Many Requests response
			api.SendResponse(w, http.StatusTooManyRequests, "Rate limit exceeded", nil, "Too many requests. Please try again later.")
			return
		}

		// If allowed, pass the request to the next handler in the chain
		next.ServeHTTP(w, r)
	}
}
