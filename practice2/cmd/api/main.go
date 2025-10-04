package main

import (
	"log"
	"net/http"
	"practice2/internal/handlers"
	"practice2/internal/middleware"
)

func main() {
	http.HandleFunc("/hello", handlers.HelloHandler)
	http.HandleFunc("/user", handlers.UserHandler)
	http.Handle("/secure", middleware.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})))
	log.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
