package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var statusUpgrader = websocket.Upgrader{}

func StatusWSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := statusUpgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	// TODO: Broadcast port I/O and route activity events here
}
