package agentdetect

import "encoding/json"

// parseJSON is a thin wrapper around json.Unmarshal for internal use.
func parseJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
