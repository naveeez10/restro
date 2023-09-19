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
	var ctx, cancel = context.WithTimeout(context.Background(),
		100*time.Second)

	matchStage := bson.D{{"$match", bson.D{{"order_id", id}}}}
	lookUpStage := bson.D{{"$lookup", bson.D{{"from", "food"},
		{"localField", "food_id"}, {"foreignField", "food_id"}, {"as",
			"food"}}}}
	unWindStage := bson.D{{"$unwind", bson.D{{"path", "$food"},
		{"preserveNullAndEmptyArrays", true}}}}

	lookUpOrderStage := bson.D{{"$lookup", bson.D{{"from", "order"},
		{"localField", "order_id"}, {"foreignField", "order_id"},
		{"as", "order"}}}}
	unWindOrderStage := bson.D{{"$unwind", bson.D{{"path", "$order"},
		{"preserveNullAndEmptyArrays", true}}}}

	lookUpTableStage := bson.D{{"$lookup", bson.D{{"from", "table"},
		{"localField", "order.table_id"}, {"foreignField",
			"order.table_id"},
		{"as", "table"}}}}
	unWindTableStage := bson.D{{"$unwind", bson.D{{"path", "$table"},
		{"preserveNullAndEmptyArrays", true}}}}

	projectStage := bson.D{
		{
			"$project", bson.D{
				{"id", 0},
				{"amount", "$food.price"},
				{"total_count", 1},
				{"food_name", "$food.name"},
				{"food_image", "$food.food_image"},
				{"table_number", "$table.table_number"},
				{"table_id", "$table.table_id"},
				{"order_id", "$order.order_id"},
				{"price", "$price.id"},
				{"quantity", "1"},
			},
		},
	}
	groupStage := bson.D{{"$group", bson.D{{"_id",
		bson.D{{"order_id", "$order_id"}, {"table_id", "$table_id"}, {"table_number", "$table_number"}}}, {"payment_due", bson.D{{"$sum", "$amount"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"order_items", bson.D{{"$push", "$$ROOT"}}}}}}

	projectStage2 := bson.D{
		{"$project", bson.D{

			{"id", 0},
			{"payment_due", 1},
			{"total_count", 1},
			{"table_number", "$_id.table_number"},
			{"order_items", 1},
		}}}

	result, err := orderItemCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage,
		lookUpStage,
		unWindStage,
		lookUpOrderStage,
		unWindOrderStage,
		lookUpTableStage,
		unWindTableStage,
		projectStage,
		groupStage,
		projectStage2})

	if err != nil {
		panic(err)
	}

	if err = result.All(ctx, &OrderItems); err != nil {
		panic(err)
	}

	defer cancel()

	return OrderItems, err
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

		orderItemstobeInserted := []interface{}{}
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
