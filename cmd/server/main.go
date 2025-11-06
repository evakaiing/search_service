package main

import (
    "searchservice/internal/handlers"
	"fmt"
    "net/http"
)


func main() {
	http.HandleFunc("/", handlers.SearchHandler)
	fmt.Println("starting server at port :8080")
	http.ListenAndServe(":8080", nil)
}