package main

import (
	"os"
	"log"
	"net/http"
	"driftGo/api/auth"
)

func main() {	
	http.HandleFunc("/auth/create", auth.SendMagicLinkCall)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
