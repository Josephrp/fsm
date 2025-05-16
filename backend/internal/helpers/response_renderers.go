package helpers

import (
	"encoding/json"
	"net/http"
)

func RenderErrorJSON(w http.ResponseWriter, statusCode uint16, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(int(statusCode))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    statusCode,
		"message": message,
	})
}
