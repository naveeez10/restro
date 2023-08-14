package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"restro/database"
	"restro/models"
	"time"
)

var orderCollection = database.OpenCollection(database.Client,
	"order")
var tableCollection = database.OpenCollection(database.Client,
	"table")

func GetOrders() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		result, err := orderCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"Error": "Error occured while listing the" +
					" order" +
					" items"})
		}
		var allOrders []bson.M
		if err = result.All(ctx, &allOrders); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allOrders)
	}
}

func GetOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		orderID := c.Param("order_id")
		var order models.Order
		err := orderCollection.FindOne(ctx,
			bson.M{"order_id": orderID}).Decode(&order)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"Error": "Error occured while finding the" +
					" order"})
		}
		c.JSON(http.StatusOK, &order)
	}
}

func CreateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order

		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)

		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(order)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		if order.Table_ID != nil {
			err := tableCollection.FindOne(ctx,
				bson.M{"table_id": order.Table_ID}).Decode(&table)

			if err != nil {
				msg := fmt.Sprintf("message : Table not found")
				defer cancel()
				c.JSON(http.StatusInternalServerError,
					gin.H{"error": msg})
				return
			}
		}

		order.Created_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))
		order.Updated_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))

		order.ID = primitive.NewObjectID()
		order.Order_ID = order.ID.Hex()

		result, insertErr := orderCollection.InsertOne(ctx, order)

		if insertErr != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "Order item was not created"})
			return
		}

		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}

func UpdateOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		var table models.Table
		var order models.Order

		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)

		var orderID string = c.Param("order_id")
		if err := c.BindJSON(&order); err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "Couldn't find any order"})
			return
		}

		var updateObj primitive.D
		if order.Table_ID != nil {
			err := menuCollection.FindOne(ctx,
				bson.M{"table_id": table.Table_ID}).Decode(&table)
			if err != nil {
				c.JSON(http.StatusInternalServerError,
					gin.H{"error": "Table not found"})

			}
			updateObj = append(updateObj,
				bson.E{"order", table.Table_ID})

		}
		order.Updated_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at", order.Updated_at})

		upsert := true

		filter := bson.M{"order_id": orderID}
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		result, err := orderCollection.UpdateOne(ctx, filter,
			bson.D{{"$st",
				updateObj}}, &opt)
		if err != nil {
			msg := fmt.Sprintf("order item update failed")
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, &result)
	}
}

func OrderItemOrderCreator(order models.Order) string {
	var _, cancel = context.WithTimeout(context.Background(),
		100*time.Second)
	order.Created_at, _ = time.Parse(time.RFC3339,
		time.Now().Format(time.RFC3339))
	order.Updated_at, _ = time.Parse(time.RFC3339,
		time.Now().Format(time.RFC3339))
	order.ID = primitive.NewObjectID()
	order.Order_ID = order.ID.Hex()
	defer cancel()
	return order.Order_ID
}
