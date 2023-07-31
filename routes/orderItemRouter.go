package routes

import (
	controller "restro/controllers"

	"github.com/gin-gonic/gin"
)

func OrderItemRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.GET("/orderItems", controller.GetOrderitems())
	incomingRoutes.GET("/orderItems/:orderItem_id", controller.GetOrderitem())
	incomingRoutes.GET("/orderItems-order/:order_id", controller.GetOrderitemsByOrder())
	incomingRoutes.POST("/orderItems", controller.CreateOrderitem())
	incomingRoutes.PATCH("/orderItems/:orderItem_id", controller.UpdateOrderitem())
}
