package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/database"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/routes"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func main() {
	fmt.Println("Hello, Golang World!")

	router:=gin.Default()

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.JSON(200,gin.H{
			"message":"Hello! Testing Gin-Gonic 🥃🍋‍🟩",
			"status_code":200,
		})
	})

	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: unable to find .env file")
	}

	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")

	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
			log.Println("Allowed Origin:", origins[i])
		}
	} else {
		origins = []string{"http://localhost:5173"}
		log.Println("Allowed Origin: http://localhost:5173")
	}

	config := cors.Config{}
	config.AllowOrigins = origins
	config.AllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"}
	//config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * time.Hour

	router.Use(cors.New(config))
	router.Use(gin.Logger())

	var client *mongo.Client = database.DBConnect() // Client obj.

	//Ping the db.
	if err:=client.Ping(context.Background(),nil);err!=nil{
		log.Fatalf("⚠️ Failed to reach server: %v",err)
	}

	defer func ()  {
		err:=client.Disconnect(context.Background())
		if err!=nil{
			log.Fatalf("⚠️ Failed to disconnect from MongoDB: %v",err)
		}
	}()

	//! routes 🛜
	routes.SetUpUnProtectedRoutes(router,client)
	routes.SetUpProtectedRoutes(router,client)

	err=router.Run()
	if err!=nil{
		fmt.Println("⚠️ ERROR starting server! ---",err)
		return
	}
}

// cd server/magikStreamMoviesServer