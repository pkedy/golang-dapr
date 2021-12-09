package dapr

import (
	"encoding/json"

	"github.com/dapr/go-sdk/service/common"
)

func DecodeTopicEvent(e *common.TopicEvent, target interface{}) (err error) {
	var payload []byte
	switch v := e.Data.(type) {
	case []byte:
		payload = v
	case string:
		payload = []byte(v)
	case map[string]interface{}:
		// This is not ideal.
		// Issue: https://github.com/dapr/go-sdk/issues/228
		payload, err = json.Marshal(v)
		if err != nil {
			return err
		}
	}
	return json.Unmarshal(payload, target)
}
