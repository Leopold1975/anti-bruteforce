package server

import (
	"net/http"
	"net/http/httptest"
	"time"

	"go.uber.org/zap"
)

func loggingMiddleware(next http.Handler, logg zap.SugaredLogger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rr := httptest.NewRecorder()

		defer func() {
			latency := time.Since(start)
			logg.Infof("%s %s %s %s %v %d %s",
				r.RemoteAddr, r.Method, r.URL.RequestURI(), r.Proto, latency, rr.Code, r.UserAgent())
		}()

		next.ServeHTTP(rr, r)

		for k, v := range rr.Header() {
			w.Header()[k] = v
		}
		w.WriteHeader(rr.Code)
		if rr.Code >= 400 && rr.Code != 429 {
			logg.Errorf("error: %s", rr.Body)
		}
		_, err := rr.Body.WriteTo(w)
		if err != nil {
			logg.Errorf("middleware write error: %w", err)
		}
	})
}
