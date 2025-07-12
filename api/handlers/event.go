package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var eventUpgrader = websocket.Upgrader{}

func EventWSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := eventUpgrader.Upgrade(w, r, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Close()
	// TODO: Broadcast device connect/disconnect and other events here
}
