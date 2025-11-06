package main

import (
    "searchservice/internal/handlers"
    "searchservice/internal/storage"
    "net/http"
)


func main() {
	http.HandleFunc("/", handler)
	fmt.Println("starting server at port :8080")
	http.ListenAndServe(":8080", nil)
}