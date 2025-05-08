package handlers

import (
	"fmt"
	"net/http"
)

func HandleHome(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Welcome to Kanban-Sim!")
}
