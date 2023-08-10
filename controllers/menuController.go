package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"time"
)

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		result, err := menuCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "Error occured while getting the menu items"})
		}
		var allMenus []bson.M
		if err = result.All(ctx, &allMenus); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allMenus)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
