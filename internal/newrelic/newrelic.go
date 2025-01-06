package newrelic

import (
	"github.com/ndn/internal/config"
	"net/http"

	"github.com/newrelic/go-agent/v3/newrelic"
)

func NewNewRelicApp(cfg *config.Config) (*newrelic.Application, error) {
	if !cfg.NewRelic.Enabled {
		return nil, nil
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(cfg.NewRelic.AppName),
		newrelic.ConfigLicense(cfg.NewRelic.LicenseKey),
		newrelic.ConfigDistributedTracerEnabled(cfg.NewRelic.DistributedTracerEnabled),
		newrelic.ConfigEnabled(true),
	)
	if err != nil {
		return nil, err
	}

	return app, nil
}

// Middleware creates a Chi middleware for New Relic instrumentation
func Middleware(app *newrelic.Application) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if app == nil {
				next.ServeHTTP(w, r)
				return
			}

			txn := app.StartTransaction(r.URL.Path)
			defer txn.End()

			w = txn.SetWebResponse(w)
			txn.SetWebRequestHTTP(r)
			r = newrelic.RequestWithTransactionContext(r, txn)

			next.ServeHTTP(w, r)
		})
	}
}
