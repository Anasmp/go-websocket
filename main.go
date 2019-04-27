package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type MessageStruct struct {
	Message string `json:"message"`
	Content string `json:"content"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan *MessageStruct)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {

	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler).Methods("GET")
	router.HandleFunc("/message", longLatHandler).Methods("POST")
	router.HandleFunc("/ws", wsHandler)
	go echo()

	log.Fatal(http.ListenAndServe(":8844", router))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "home")
}

func writer(coord *MessageStruct) {
	broadcast <- coord
}

func longLatHandler(w http.ResponseWriter, r *http.Request) {
	var coordinates MessageStruct
	if err := json.NewDecoder(r.Body).Decode(&coordinates); err != nil {
		log.Printf("ERROR: %s", err)
		http.Error(w, "Bad request", http.StatusTeapot)
		return
	}
	defer r.Body.Close()
	go writer(&coordinates)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// register client
	clients[ws] = true
}

func echo() {
	for {
		val := <-broadcast
		latlong := fmt.Sprintf("%s %s", val.Message, val.Content)

		// send to every client that is currently connected
		for client := range clients {
			err := client.WriteMessage(websocket.TextMessage, []byte(latlong))
			if err != nil {
				log.Printf("Websocket error: %s", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
