package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/kuromii5/chat-bot-auth-service/internal/domain"
	"github.com/kuromii5/chat-bot-shared/wrapper"
)

type jailEntry struct {
	failures int
	jailedAt time.Time
	jailed   bool
}

type IPJail struct {
	mu          sync.Mutex
	entries     map[string]*jailEntry
	maxFailures int
	jailDur     time.Duration
}

func NewIPJail(maxFailures int, jailDur time.Duration) *IPJail {
	j := &IPJail{
		entries:     make(map[string]*jailEntry),
		maxFailures: maxFailures,
		jailDur:     jailDur,
	}
	go j.cleanup()
	return j
}

// cleanup removes expired jail entries every 5 minutes to prevent memory leak.
func (j *IPJail) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		j.mu.Lock()
		for ip, entry := range j.entries {
			if entry.jailed && time.Since(entry.jailedAt) > j.jailDur {
				delete(j.entries, ip)
			}
		}
		j.mu.Unlock()
	}
}

func (j *IPJail) isJailed(ip string) bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	entry, ok := j.entries[ip]
	if !ok {
		return false
	}
	if entry.jailed && time.Since(entry.jailedAt) > j.jailDur {
		delete(j.entries, ip)
		return false
	}
	return entry.jailed
}

func (j *IPJail) recordFailure(ip string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	entry, ok := j.entries[ip]
	if !ok {
		entry = &jailEntry{}
		j.entries[ip] = entry
	}
	entry.failures++
	if entry.failures >= j.maxFailures {
		entry.jailed = true
		entry.jailedAt = time.Now()
	}
}

func (j *IPJail) recordSuccess(ip string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.entries, ip)
}

// Middleware checks if an IP is jailed before the request and records
// failed login attempts (401) after the handler responds.
func (j *IPJail) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)

		if j.isJailed(ip) {
			wrapper.WrapError(w, r, domain.ErrIPJailed)
			return
		}

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		switch rec.status {
		case http.StatusUnauthorized:
			j.recordFailure(ip)
		case http.StatusOK:
			j.recordSuccess(ip)
		}
	})
}

// statusRecorder wraps ResponseWriter to capture the response status code.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

func (sr *statusRecorder) Write(b []byte) (int, error) {
	if sr.status == 0 {
		sr.status = http.StatusOK
	}
	return sr.ResponseWriter.Write(b)
}

// clientIP extracts the real client IP from RemoteAddr.
// RealIP middleware runs before route-level middleware, so RemoteAddr is already correct.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
