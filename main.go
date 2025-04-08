package main

//go:generate go run generate.go

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/rvodden/teams/internal/generated_data"
)

// getAlbums responds with the list of all albums as JSON.
func getPeople(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, generated_data.People)
}

func getTeams(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, generated_data.Teams)
}

func main() {
	router := gin.Default()
	router.GET("/people", getPeople)
	router.GET("/teams", getTeams)

	err := router.Run("localhost:8080")
	if err != nil {
	}
}
