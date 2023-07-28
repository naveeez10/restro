package main

import(
	"os"
	"restro/database"
	"restro/routes"
)

func main(
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())
)
