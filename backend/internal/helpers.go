package internal

import "encoding/json"

func ToJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
