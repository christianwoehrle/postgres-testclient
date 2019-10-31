package prometheus_cw

import "encoding/json"

// taken from here: https://github.com/prometheus/client_golang/blob/master/api/prometheus/v1/api.go
type ApiResponse struct {
	Status    string          `json:"status"`
	Data      json.RawMessage `json:"data"`
	ErrorType ErrorType       `json:"errorType"`
	Error     string          `json:"error"`
	Warnings  []string        `json:"warnings,omitempty"`
}

// queryResult contains result data for a query.
type queryResult struct {
	Type   ValueType   `json:"resultType"`
	Result interface{} `json:"result"`

	// The decoded value.
}

type ErrorType string

// HealthStatus models the health status of a scrape target.
type HealthStatus string

// https://github.com/prometheus/common/blob/master/model/value.go

type ValueType int

const (
	ValNone ValueType = iota
	ValScalar
	ValVector
	ValMatrix
	ValString
)
