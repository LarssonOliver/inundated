package middleware

import (
	"errors"
	"log/slog"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/larssonoliver/inundated/internal/logging"
	"github.com/larssonoliver/inundated/internal/model"
	"github.com/larssonoliver/inundated/internal/service"
)

const (
	AppPasswordFailureLimit  = 10
	AppPasswordFailureWindow = time.Minute
)

// AppPasswordAuth authenticates HTTP Basic requests using an app password as
// the password; the username is ignored. requireOwner rejects app passwords
// that belong to no user.
func AppPasswordAuth(svc service.AppPasswordService, requireOwner bool) func(http.Handler) http.Handler {
	failures := httprate.NewRateLimiter(AppPasswordFailureLimit, AppPasswordFailureWindow)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, _ := clientIPKey(r)

			// Status only reports "over" once the rate exceeds the limit, which
			// would allow one attempt too many; compare the rate ourselves.
			_, rate, err := failures.Status(key)
			if err != nil {
				respondServiceUnavailable(w)
				return
			}
			if int(math.Round(rate)) >= AppPasswordFailureLimit {
				limitJSON429(w, r)
				return
			}

			_, token, ok := r.BasicAuth()
			if !ok || token == "" {
				rejectAppPassword(w, r, failures, key)
				return
			}

			appPassword, err := svc.AuthenticateAppPassword(r.Context(), token)
			if errors.Is(err, model.ErrNotFound) {
				rejectAppPassword(w, r, failures, key)
				return
			}
			if err != nil {
				respondServiceUnavailable(w)
				return
			}
			if appPassword.UserId == nil && requireOwner {
				rejectAppPassword(w, r, failures, key)
				return
			}

			ctx := r.Context()
			if appPassword.UserId != nil {
				ctx = model.SetUserInContext(ctx, model.User{Id: *appPassword.UserId})
				ctx = logging.ContextWith(ctx, slog.String("user_id", appPassword.UserId.String()))
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func rejectAppPassword(w http.ResponseWriter, r *http.Request, failures *httprate.RateLimiter, key string) {
	failures.OnLimit(w, r, key)
	w.Header().Set("WWW-Authenticate", `Basic realm="inundated"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}
