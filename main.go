package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/labstack/echo"
)

const (
	MAXRETRIES  = 10
	RETYBACKOFF = 500
	PORT        = "8000"
)

func main() {
	var db *gorm.DB
	var err error
	for retries := 0; retries < MAXRETRIES; retries++ {
		db, err = gorm.Open(postgres.New(postgres.Config{
			DSN:                  os.Getenv("DSN"),
			PreferSimpleProtocol: true, // disables implicit prepared statement usage
		}), &gorm.Config{})
		if err == nil {
			break
		} else {
			time.Sleep(RETYBACKOFF * time.Millisecond)
		}
	}
	if err != nil {
		log.Fatalf("failed to connect to Postgres: %v", err)
	}

	// Create an API handler which serves data from Postgres.
	e := echo.New()

	err = db.AutoMigrate(&Function{})
	if err != nil {
		log.Panicln("Error running db migration")
	}

	e.Use(bindDb(db))

	e.GET("/ping", func(c echo.Context) error {
		pingResponse := PingResponse{Ok: true}
		return c.JSON(http.StatusOK, pingResponse)
	})

	e.POST("/functions", CreateFunctionHandler)
	e.POST("/upload", UploadHandler)
	e.POST("/invoke/:name", InvokeHandler)
	e.GET("/functions", GetFunctionsHandler)
	e.GET("/functions/:uuid", GetFunctionHandler)
	e.PUT("/functions/:uuid", UpdateFunctionHandler)
	e.DELETE("/functions/:uuid", DeleteFunctionHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = PORT
	}
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", port)))
}

// Bind the db as a middleware
func bindDb(db *gorm.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("db", db)
			return next(c)
		}
	}
}
