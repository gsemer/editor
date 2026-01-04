package main

import (
	"editor/application"
	"editor/domain"
	"editor/infrastructure"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		DB:       0,
		Password: "",
	})
	bus := application.NewRedisBus(rdb)

	documentID := uuid.New().String()
	documentContent := "Welcome to 2026 Gophers!"
	document := domain.NewDocument(documentID, documentContent)

	hub := application.NewEditorHub(document)
	go hub.Run(bus)

	handler := infrastructure.NewWSServer(hub)
	http.HandleFunc("/ws", handler.WSHandler)

	log.Println("Listening to port 3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
