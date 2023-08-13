package controller

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"restro/models"
	"time"
)

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		menuId := c.Param("menu_id")
		var menu models.Menu
		err := foodCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while fetching the menu item"})
		}
		c.JSON(http.StatusOK, menu)
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
		var menu models.Menu
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(menu)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		menu.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

		menu.ID = primitive.NewObjectID()
		menu.Menu_ID = menu.ID.Hex()

		result, insertErr := menuCollection.InsertOne(ctx, menu)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't add the menu"})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
		defer cancel()
	}
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var menuID string = c.Param("menu_id")
		filter := bson.M{"manuu_id": menuID}

		var updateObj primitive.D

		if menu.Start_Date != nil && menu.End_Date != nil {
			if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now()) {
				msg := "kindly retype the time"
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				defer cancel()
				return
			}
			updateObj = append(updateObj, bson.E{"start_date", menu.Start_Date})
			updateObj = append(updateObj, bson.E{"end_date", menu.End_Date})

			if menu.Name != "" {
				updateObj = append(updateObj, bson.E{"name", menu.Name})
			}
			if menu.Category != "" {
				updateObj = append(updateObj, bson.E{"category", menu.Category})
			}
			menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			updateObj = append(updateObj, bson.E{"updated_at", menu.Updated_at})

			upsert := true

			opt := options.UpdateOptions{
				Upsert: &upsert,
			}

			result, err := menuCollection.UpdateOne(ctx, filter, bson.D{
				{"$set", updateObj},
			}, &opt)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Menu Update Failed"})
			}
			defer cancel()
			c.JSON(http.StatusOK, result)

		}
	}
}
