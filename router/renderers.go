package router

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// func renderJSON[T any](w http.ResponseWriter, r *http.Request, status int, obj T) error {
func RenderJSON[T any](w http.ResponseWriter, status int, obj T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(obj); err != nil {
		return fmt.Errorf("encoder: %w", err)
	}

	return nil
}

// data, err := readJSON[MyStructType](r)
func ReadJSON[T any](req *http.Request) (T, error) {
	var obj T
	if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
		return obj, fmt.Errorf("decoder: %w", err)
	}

	return obj, nil
}

func RenderHTML(w http.ResponseWriter, status int, s string) error {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	_, err := w.Write([]byte(s))
	return err
}
