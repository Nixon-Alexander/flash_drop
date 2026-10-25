package routes

import (
	"flash_drop/controller"
	"flash_drop/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ControllerHandler struct {
	r           *gin.Engine
	pc          *controller.ProductController
	uc          *controller.UserController
	cc          *controller.CategoryController
	cart_c      *controller.CartController
	checkout_c  *controller.CheckoutController
	flashSale_c *controller.FlashSaleController
	payment_c   *controller.PaymentController
}

func InitRoutes(
	r *gin.Engine,
	pc *controller.ProductController,
	uc *controller.UserController,
	cc *controller.CategoryController,
	cart_c *controller.CartController,
	checkout_c *controller.CheckoutController,
	flashSale_c *controller.FlashSaleController,
	payment_c *controller.PaymentController,
) *ControllerHandler {
	return &ControllerHandler{
		r:           r,
		pc:          pc,
		uc:          uc,
		cc:          cc,
		cart_c:      cart_c,
		checkout_c:  checkout_c,
		flashSale_c: flashSale_c,
		payment_c:   payment_c,
	}
}

func (ch *ControllerHandler) StartRoutes(router *gin.Engine) {
	v1 := ch.r.Group("/api/v1")
	v1.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// User
	v1.POST("/signup", ch.uc.Register)
	v1.POST("/login", ch.uc.LogIn)
	v1.POST("/verify_token", ch.uc.VerifyToken)
	v1.POST("/payment/notification", ch.payment_c.Notification)

	// User Product
	product := v1.Group("/product")
	product.GET("/", ch.pc.GetAllProduct)

	// User Cart
	cart := v1.Group("/cart")
	cart.Use(middleware.AuthMiddleware("customer"))
	cart.GET("/", ch.cart_c.GetItems)
	cart.POST("/", ch.cart_c.AddItems)
	cart.PUT("/", ch.cart_c.UpdateItems)

	// User Checkout
	checkout := v1.Group("/checkout")
	checkout.Use(middleware.AuthMiddleware("customer"))
	checkout.POST("/", ch.checkout_c.BuyItems)

	// User Flash Sale
	userFlashSales := v1.Group("/flash_sales")
	userFlashSales.Use(middleware.AuthMiddleware("customer"))
	userFlashSales.POST("/checkout", ch.checkout_c.BuyItems)

	// Admin
	adminRoutes := v1.Group("/admin")
	adminRoutes.Use(middleware.AuthMiddleware("admin"))
	// -- Product
	product = adminRoutes.Group("/product")
	product.POST("/", ch.pc.CreateProduct)
	product.DELETE("/:id", ch.pc.DeleteProduct)
	product.PUT("/:id", ch.pc.EditProduct)

	// -- Category
	category := adminRoutes.Group("/category")
	category.POST("/", ch.cc.CreateCategory)
	category.DELETE("/:id", ch.cc.DeleteCategory)
	category.PUT("/:id", ch.cc.EditCategory)
	category.GET("/", ch.cc.GetAllCategory)

	// -- Flash Sales
	flashSales := adminRoutes.Group("/flash_sales")
	flashSales.POST("/", ch.flashSale_c.CreateEvent)
	flashSales.GET("/", ch.flashSale_c.GetAllEvent)
	flashSales.GET("/:id", ch.flashSale_c.GetEventById)
	flashSales.PATCH("/:id/status", ch.flashSale_c.UpdateStatus)
	flashSales.DELETE("/:id", ch.flashSale_c.DeleteEvent)
}
