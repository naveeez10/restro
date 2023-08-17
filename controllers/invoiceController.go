package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"restro/database"
	"restro/models"
	"time"
)

type InvoiceViewFormat struct {
	Invoice_id       string
	Payment_method   string
	Order_id         string
	Payment_Status   *string
	Payment_due      interface{}
	Table_number     interface{}
	Payment_due_date time.Time
	Order_details    interface{}
}

var invoiceCollection *mongo.Collection = database.OpenCollection(
	database.Client,
	"invoice")

func GetInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		var invoice models.Invoice
		invoiceId := c.Param("invoice_id")
		err := invoiceCollection.FindOne(ctx,
			bson.M{"invoice_id": invoiceId}).Decode(&invoice)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "error occured while getting the" +
					" required invoice"})
			return
		}
		var invoiceView InvoiceViewFormat
		allOrderItems, err := ItemsByOrder(invoice.Order_ID)
		invoiceView.Order_id = invoice.Order_ID
		invoiceView.Payment_due_date = invoice.Payment_due_date
		invoiceView.Payment_method = "null"

		if invoice.Payment_Method != nil {
			invoiceView.Payment_method = *invoice.Payment_Method
		}
		invoiceView.Invoice_id = invoice.Invoice_ID
		invoiceView.Payment_Status = *&invoice.Payment_Status
		invoiceView.Payment_due = allOrderItems[0]["payment_due"]
		invoiceView.Table_number = allOrderItems[0]["table_number"]
		invoiceView.Order_details = allOrderItems[0]["order_details"]

		c.JSON(http.StatusOK, invoiceView)
	}
}

func GetInvoices() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		result, err := invoiceCollection.Find(ctx, bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "Couldn't list the invoices"})
			return
		}
		var allInvoices []bson.M
		if err := result.All(ctx, &allInvoices); err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, allInvoices)
	}
}

func CreateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)
		var invoice models.Invoice
		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": err.Error()})
			return
		}
		var order models.Order

		err := orderCollection.FindOne(ctx,
			bson.M{"order_id": invoice.Order_ID}).Decode(&order)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("couldn't find any order with given" +
				" order id")
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": msg})
			return
		}
		status := "PENDING"
		if invoice.Payment_Status == nil {
			invoice.Payment_Status = &status
		}
		invoice.Payment_due_date, _ = time.Parse(time.RFC3339,
			time.Now().AddDate(0, 0, 1).Format(time.RFC3339))
		invoice.Created_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))
		invoice.Updated_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))
		invoice.ID = primitive.NewObjectID()
		invoice.Invoice_ID = invoice.ID.Hex()

		validationErr := validate.Struct(invoice)
		if validationErr != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": validationErr.Error()})
			return
		}
		result, err := invoiceCollection.InsertOne(ctx, invoice)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": "Invoice item was not created"})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, result)
	}
}

func UpdateInvoice() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(),
			100*time.Second)

		var invoice models.Invoice
		invoiceID := c.Param("invoice_id")

		if err := c.BindJSON(&invoice); err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": err.Error()})
			return
		}
		filter := bson.M{"invoice_id": invoiceID}
		var updateObj primitive.D

		if invoice.Payment_Method != nil {
			updateObj = append(updateObj,
				bson.E{"payment_method", invoice.Payment_Method})
		}
		if invoice.Payment_Status != nil {
			updateObj = append(updateObj, bson.E{"payment_status",
				invoice.Payment_Status})
		}
		invoice.Updated_at, _ = time.Parse(time.RFC3339,
			time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{"updated_at",
			invoice.Updated_at})

		upsert := true
		opt := options.UpdateOptions{
			Upsert: &upsert,
		}
		status := "PENDING"
		if invoice.Payment_Status == nil {
			invoice.Payment_Status = &status
		}
		result, err := invoiceCollection.UpdateOne(ctx, filter,
			bson.D{{"$set",
				updateObj}}, &opt)

		if err != nil {
			msg := fmt.Sprintf("Invoice item update failed")
			c.JSON(http.StatusInternalServerError,
				gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)
	}
}
