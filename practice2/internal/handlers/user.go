package handlers

import (
	"encoding/json"
	_ "encoding/json"
	_ "fmt"
	"net/http"
	_ "net/http"
	"strconv"
	_ "strconv"
)

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")

	switch id, err := strconv.Atoi(idStr); {
	case idStr == "":
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
	case err != nil:
		http.Error(w, `{"error": "invalid id"}`, http.StatusBadRequest)
	default:
		respons := map[string]interface{}{
			"message": "Hello " + strconv.Itoa(id),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(respons)
	}
}

type UserRequest struct {
	Name string `json:"name"`
}

func UserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req UserRequest
	err := json.NewDecoder(r.Body).Decode(&req)

	switch {
	case err != nil:
		http.Error(w, `{"error": "invalid JSON"}`, http.StatusBadRequest)

	case req.Name == "":
		http.Error(w, `{"error": "invalid name"}`, http.StatusBadRequest)

	default:
		response := map[string]interface{}{
			"created": req.Name,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}
