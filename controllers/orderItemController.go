package controller

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"restro/database"
	"restro/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type orderItemPack struct {
	Table_id   *string
	OrderItems []models.OrderItem
}

var orderItemCollection *mongo.Collection = database.OpenCollection(database.
	Client,
	"orderItem")

func GetOrderItemsByOrder() gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("order_id")
		allOrderItems, err := ItemsByOrder(orderID)

		if err != nil {
			c.JSON(500,
				gin.H{"error": "error occurred while listing the" +
					" order items with given order ID"})
			return
		}
		c.JSON(200, &allOrderItems)
	}
}

func GetOrderItems() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		result, err := orderItemCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(500,
				gin.H{"error": "Error occurred while fetching the" +
					" order items"})
			return
		}
		var allOrderItems []bson.M
		if err = result.All(ctx, &allOrderItems); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, &allOrderItems)
	}
}

func ItemsByOrder(id string) (OrderItems []primitive.M, err error) {

}

func GetOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		var orderItem models.OrderItem
		orderItemId := c.Param("order_item_id")
		err := orderItemCollection.FindOne(ctx,
			bson.M{"order_item_id": orderItemId}).Decode(&orderItem)
		defer cancel()
		if err != nil {
			c.JSON(500,
				gin.H{"error": "couldn't find any order item with" +
					" given ID"})
			return
		}
		c.JSON(http.StatusOK, &orderItem)
	}
}

func UpdateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		var orderItem models.OrderItem

		orderItemId := c.Param("order_item_id")
		filter := bson.M{"order_item_id": orderItemId}
		var updateObj primitive.D
		if orderItem.Unit_price != nil {
			updateObj = append(updateObj, bson.E{"unit_price",
				orderItem.Unit_price})
		}
		if orderItem.Quantity != nil {
			updateObj = append(updateObj, bson.E{"quantity",
				orderItem.Quantity})
		}
		if orderItem.Food_id != nil {
			updateObj = append(updateObj, bson.E{"food_id",
				orderItem.Updated_at})
		}

		orderItem.Updated_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at",
			orderItem.Updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := orderItemCollection.UpdateOne(
			ctx, filter, bson.D{{"$set", updateObj}}, &opt)

		if err != nil {
			c.JSON(500,
				gin.H{"error": "Couldn't update the orderItem with" +
					" the specified ID"})
			return
		}
		defer cancel()
		c.JSON(200, &result)
	}
}

func CreateOrderItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		var orderItempack orderItemPack
		var order models.Order

		if err := c.BindJSON(&orderItempack); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		order.Order_date, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))

		orderItemstobeInserted := []interface{}
		order.Table_ID = orderItempack.Table_id
		order_id := OrderItemOrderCreator(order)

		for _, orderItem := range orderItempack.OrderItems {
			orderItem.Order_id = order_id
			validationErr := validate.Struct(order)
			if validationErr != nil {
				c.JSON(500, gin.H{"error": validationErr.Error()})
				return
			}
			orderItem.ID = primitive.NewObjectID()
			orderItem.Created_at, _ = time.Parse(time.RFC3339,
				time.Now().Format(time.RFC3339))
			orderItem.Updated_at, _ = time.Parse(time.RFC3339,
				time.Now().Format(time.RFC3339))
			orderItem.Order_item_id = orderItem.ID.Hex()
			var num = toFixed(*orderItem.Unit_price, 2)
			orderItem.Unit_price = &num
			orderItemstobeInserted = append(orderItemstobeInserted,
				orderItem)
		}
		insertedOrderItems, err := orderItemCollection.InsertMany(
			ctx,
			orderItemstobeInserted)
		defer cancel()

		if err != nil {
			log.Fatal(err)
		}

		c.JSON(200, &insertedOrderItems)
	}
}
