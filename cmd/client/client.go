package main

import (
	"fmt"
	"log"
	"searchservice/internal/models"
	"searchservice/pkg/client"
)

func main() {
	searchClient := client.SearchClient{
		AccessToken: "token1",
		URL:         "http://localhost:8080",
	}

	resp, err := searchClient.FindUsers(models.SearchRequest{
		Limit:      10,
		Offset:     0,
		OrderField: "Name",
		OrderBy:    models.OrderByAsc,
	})
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("found %d users\n", len(resp.Users))
	for _, user := range resp.Users {
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", user.ID, user.Name, user.Age)
	}
}
