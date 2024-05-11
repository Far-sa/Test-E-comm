package main

import (
	"auth-svc/adapters/config"
	"auth-svc/adapters/delivery"
	"auth-svc/adapters/messaging"
	"auth-svc/adapters/repository/db"
	"auth-svc/adapters/repository/migrator"
	"auth-svc/adapters/repository/postgres"
	"auth-svc/internal/service"
	"fmt"
	"log"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	viper.ReadInConfig()
}

func main() {

	configAdapter, err := config.NewViperAdapter()
	if err != nil {
		fmt.Println("failed to load configuration", err)
	}

	dbPool, err := db.GetConnectionPool(configAdapter) // Use dedicated function (if using db package)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer dbPool.Close() // Close the pool when done (consider connection pool management)

	authRepo := postgres.NewAuthRepository(dbPool)

	mgr := migrator.NewMigrator(dbPool, "database/migrations")
	mgr.MigrateUp()

	log.Println("Migrations completed successfully!")

	//connectionString := "amqp://guest:guest@localhost:5672/"
	eventPublisher, err := messaging.NewRabbitClient(configAdapter)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}

	authSvc := service.NewAuthService(configAdapter, authRepo, eventPublisher)

	authHandler := delivery.NewAuthHandler(authSvc)

	// Initialize Echo
	e := echo.New()
	// e.Use(middleware.Logger())
	// e.Use(middleware.Recover())

	e.POST("/login", authHandler.Login)
	// e.GET("/revoke-token", authHandler.RevokeToken)

	// 	// Start server
	e.Logger.Fatal(e.Start(":5001"))

}
