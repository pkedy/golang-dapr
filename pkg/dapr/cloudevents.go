package dapr

import (
	"encoding/json"
	"time"
)

// CloudEvents specification can be found at
// https://github.com/cloudevents/spec/blob/v1.0.1/spec.md
type CloudEvent struct {
	ID              string          `json:"id"`
	Source          string          `json:"source"`
	SpecVersion     string          `json:"specversion"`
	Type            string          `json:"type"`
	DataContentType string          `json:"datacontenttype,omitempty"`
	DataSchema      string          `json:"dataschema,omitempty"`
	Subject         string          `json:"subject,omitempty"`
	Time            time.Time       `json:"time,omitempty"`
	Data            json.RawMessage `json:"data"`
}
