package helpers

import "net/http"

func HeaderJSON(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
}

func HeaderText(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html")
}
