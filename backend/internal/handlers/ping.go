package handlers

import (
	"fmt"
	"net/http"
)

func HandlePing(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	fmt.Fprintln(w, "Backend running")
}
