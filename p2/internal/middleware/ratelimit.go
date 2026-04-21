package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Tanveer-rajpurohit/p2/internal/auth"
	"github.com/Tanveer-rajpurohit/p2/internal/utils"
	"golang.org/x/time/rate"
)

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type Store struct {
	mu       sync.Mutex
	limiters map[string]*entry
	interval time.Duration
	burst    int
}

func NewStore(refill time.Duration, burst int) *Store {
	s := &Store{
		limiters: make(map[string]*entry),
		interval: refill,
		burst:    burst,
	}

	go s.cleanupLoop()

	return s
}

func (s *Store) get(userId string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, exists := s.limiters[userId]
	if !exists {
		e = &entry{
			limiter:  rate.NewLimiter(rate.Every(s.interval), s.burst),
			lastSeen: time.Now(),
		}
		s.limiters[userId] = e
	}

	e.lastSeen = time.Now()
	return e.limiter
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for userId, e := range s.limiters {
			if time.Since(e.lastSeen) > 5*time.Minute {
				delete(s.limiters, userId)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Store) Limit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := r.Context().Value(auth.ClaimsKey).(*auth.Claims)

		userID := ""
		if claims != nil {
			userID = claims.UserID
		}

		// Fallback: use IP for unauthenticated routes
		if userID == "" {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				userID = r.RemoteAddr
			} else {
				userID = ip
			}
		}

		if !s.get(userID).Allow() {
			utils.ResponseWithError(w, http.StatusTooManyRequests, "too many requests")
			return
		}
		next.ServeHTTP(w, r)
	})
}
