package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("campaign events API listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", &API{store: NewStore()}); err != nil {
		log.Fatal(err)
	}
}
