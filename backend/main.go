package main

import (
	"context"
	"errors"
	service "flash_drop/Service"
	"flash_drop/controller"
	"flash_drop/database"
	"flash_drop/helper"
	"flash_drop/middleware"
	"flash_drop/repository"
	"flash_drop/routes"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Init logger
	logger := helper.CreateLogger()
	defer logger.Sync()

	// Load .env
	err := godotenv.Load()
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
	paymentController := controller.InitPaymentController(logger, checkoutService)

	flashSaleRepo := repository.InitFlashSaleRepository(db)
	flashSaleController := controller.InitFlashSaleController(logger, flashSaleRepo)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		checkoutService.StartExpiryWorker(ctx, time.Minute)
	}()

	router := gin.Default()
	router.MaxMultipartMemory = 8 << 20
	_ = router.SetTrustedProxies(nil)
	router.Static("/uploads", "./uploads")
	router.Use(middleware.CorsMiddleware())

	// Start Serve
	routeHandler := routes.InitRoutes(router, productController, userController, categoryController, cartController, checkoutController, flashSaleController, paymentController)
	routeHandler.StartRoutes(router)

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%s", os.Getenv("APP_HOST"), os.Getenv("APP_PORT")),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal("server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	wg.Wait()
}
