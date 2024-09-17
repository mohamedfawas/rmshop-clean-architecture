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
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter := rate.NewLimiter(i.r, i.b)

	i.ips[ip] = limiter

	return limiter
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	limiter, exists := i.ips[ip]

	if !exists {
		i.mu.Unlock()
		return i.AddIP(ip)
	}

	i.mu.Unlock()

	return limiter
}

func RateLimitMiddleware(next http.HandlerFunc, limiter *IPRateLimiter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		if !limiter.GetLimiter(ip).Allow() {
			api.SendResponse(w, http.StatusTooManyRequests, "Rate limit exceeded", nil, "Too many requests. Please try again later.")
			return
		}

		next.ServeHTTP(w, r)
	}
}
