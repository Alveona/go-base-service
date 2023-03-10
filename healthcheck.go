package base_swagger_service

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

const Ok = "ok"

type Healthcheck struct {
	Version    string
	ServerName string
	AppName    string
	Strategy   StrategyError
	checkers   []Checker
}

type Checker func() CheckerResult

type CheckerResult struct {
	Service string `json:"service"`
	Status  bool   `json:"status"`
}

func (h *Healthcheck) Check() ([]byte, bool, error) {
	// run all checkers
	var checks []CheckerResult
	for _, c := range h.checkers { // @todo make goroutines
		checks = append(checks, c())
	}

	hasProblem := h.Strategy.setError(checks)

	// prepare result
	result := map[string]interface{}{
		"server":      h.ServerName,
		"version":     h.Version,
		"application": h.AppName,
	}
	if len(checks) > 0 {
		result["checkResultList"] = checks
	}

	// convert to JSON
	if jsn, err := json.Marshal(result); err == nil {
		return jsn, hasProblem, nil
	} else {
		return nil, hasProblem, err
	}
}

func (h *Healthcheck) AddChecker(c Checker) {
	h.checkers = append(h.checkers, c)
}

func DbChecker(name string, db *sql.DB) Checker {
	return func() CheckerResult {
		if err := db.Ping(); err == nil {
			return CheckerResult{name, true}
		} else {
			return CheckerResult{name, false}
		}
	}
}

func Handler(hc *Healthcheck) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "application/json; charset=utf-8")

		// if have not strategy => add default strategy
		if hc.Strategy == nil {
			hc.Strategy = StrategyErrorOne{}
		}

		res, hasProblem, err := hc.Check()

		// if has problem with server or has error => 500
		if hasProblem == true || err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		if err == nil {
			w.Write(res)
		} else {
			w.Write([]byte(err.Error()))
		}
	})
}
