package main

import (
	service "flash_drop/Service"
	"flash_drop/controller"
	"flash_drop/database"
	"flash_drop/helper"
	"flash_drop/repository"
	"flash_drop/routes"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Init logger
	logger := helper.CreateLogger()
	defer logger.Sync()

	// Load .env
	err := godotenv.Load("local.env")
	if err != nil {
		logger.Error(err.Error())
		return
	}

	// Init DB
	db, err := database.InitDatabase()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	database.RunMigration(logger)

	defer db.Close()

	// Injection
	authRepo := repository.InitAuthRepository(db)
	userController := controller.InitUserController(logger, authRepo)

	categoryRepo := repository.InitCategoryRepository(db)
	categoryController := controller.InitCategoryController(logger, categoryRepo)

	productRepo := repository.InitProductRepository(db)
	productController := controller.InitProductController(logger, productRepo)

	cartRepo := repository.InitCartRepository(db)
	cartController := controller.InitCartController(logger, cartRepo)

	orderRepo := repository.InitOrderRepository(db)
	paymentRepo := repository.InitPaymentRepository(db)

	paymentService := service.InitPaymentService(logger, paymentRepo)
	checkoutService := service.InitCheckoutService(db, logger, orderRepo, paymentService)
	checkoutController := controller.InitCheckoutController(logger, checkoutService)

	flashSaleRepo := repository.InitFlashSaleRepository(db)
	flashSaleController := controller.InitFlashSaleController(logger, flashSaleRepo)

	// Start Serve
	router := gin.Default()
	routeHanlder := routes.InitRoutes(router, productController, userController, categoryController, cartController, checkoutController, flashSaleController)
	routeHanlder.StartRoutes(router)
	app := fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT"))
	router.Run(app)
}
