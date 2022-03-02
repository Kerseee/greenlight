package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"greenlight.kerseeehuang.com/internal/data"
	"greenlight.kerseeehuang.com/internal/validator"
)

// recoverPanic is a middleware. It recovers from the panic in the handler next
// and send connection-close response to clients.
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Recover from the panic.
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// rateLimit is a middlerware that wraps the handler next with a map of ip-based rate limiter.
// It also automatically deletes limiters of clients that are not seen for a long time.
func (app *application) rateLimit(next http.Handler) http.Handler {
	// client stores the limiter of a client and the last seem time of it.
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Create a map of rate limiters.
	var (
		mu        sync.Mutex
		clients   = make(map[string]*client)
		stayLimit = 3 * time.Minute
	)

	// Remove the limiters of clients that are not seen for a long time every minute.
	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			// Remove limiters.
			for ip, c := range clients {
				if time.Since(c.lastSeen) > stayLimit {
					delete(clients, ip)
				}
			}

			mu.Unlock()
		}
	}()

	// Wrap next with limiter.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract the ip from the request r.
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()
		// Get the limiter of this ip.
		// If this ip is not in clients, then create a limiter for it.
		if _, ok := clients[ip]; !ok {
			clients[ip] = &client{
				limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
			}
		}
		// Update lastSeen of this client.
		clients[ip].lastSeen = time.Now()

		// Take a token. If no token is available, drop this request and inform the client.
		if !clients[ip].limiter.Allow() {
			app.rateLimitExceededResponse(w, r)
			mu.Unlock()
			return
		}

		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

// authenticate is a middleware that authenticates the user in the request.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Indicates in header that the response may vary based on authentication.
		w.Header().Set("Vary", "Authorization")

		// Get the authorization header.
		authorizationHeader := r.Header.Get("Authorization")

		// Directly call next the user is no authorized.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// Varify the authorization token.
		authTokenParts := strings.Split(authorizationHeader, " ")
		if len(authTokenParts) != 2 || authTokenParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Extract the token.
		token := authTokenParts[1]

		// Validate the token.
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get the user from DB by this token.
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Put user into the request context.
		r = app.contextSetUser(r, user)

		// Call the next handler.
		next.ServeHTTP(w, r)
	})
}
