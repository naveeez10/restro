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

func GetTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		tableID := c.Param("table_id")
		var table models.Table
		err := tableCollection.FindOne(ctx,
			bson.M{"table_id": tableID}).Decode(&table)
		defer cancel()
		if err != nil {
			c.JSON(500,
				gin.H{"error": "Couldn't find any table with given" +
					" table ID"})
			return
		}
		c.JSON(200, &table)
	}
}

func GetTables() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		var tables []bson.M
		result, err := tableCollection.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(500,
				gin.H{"error": "Couldn't list all the tables"})
			return
		}
		if err = result.All(ctx, tables); err != nil {
			log.Fatal(err)
		}
		c.JSON(200, &tables)
	}
}

func CreateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)

		var table models.Table

		if err := c.BindJSON(&table); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(table)
		if validationErr != nil {
			c.JSON(400, gin.H{"error": validationErr.Error()})
			return
		}

		table.Created_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))
		table.Updated_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))

		table.ID = primitive.NewObjectID()
		table.Table_ID = table.ID.Hex()

		result, insertErr := tableCollection.InsertOne(ctx, table)
		if insertErr != nil {
			c.JSON(400, gin.H{"error": "Couldn't insert the table"})
			return
		}
		defer cancel()
		c.JSON(200, &result)
	}
}

func UpdateTable() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		var table models.Table
		tableID := c.Param("table_id")
		if err := c.BindJSON(&table); err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if table.Number_of_guests != nil {
			updateObj = append(updateObj,
				bson.E{"number_of_guests", table.Number_of_guests})
		}

		if table.Table_Number != nil {
			updateObj = append(updateObj, bson.E{"table_number",
				table.Table_Number})
		}

		table.Updated_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		filter := bson.M{"table_id": tableID}

		result, err := tableCollection.UpdateOne(ctx, filter, bson.D{{
			"#set", updateObj,
		},
		}, &opt)

		if err != nil {
			c.JSON(500, gin.H{"error": "Couldn't update the table" +
				" information"})
		}
		defer cancel()
		c.JSON(200, &result)
	}
}
