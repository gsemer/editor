package main

import (
	"editor/application"
	"editor/domain"
	"editor/infrastructure"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func main() {
	documentID := uuid.New().String()
	documentContent := "Welcome to 2026 Gophers!"
	document := domain.NewDocument(documentID, documentContent)

	hub := application.NewEditorHub(document)
	go hub.Run()

	handler := infrastructure.NewWSServer(hub)
	http.HandleFunc("/ws", handler.WSHandler)

	log.Println("Listening to port 3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
