package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/controllers"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetUpProtectedRoutes(router *gin.Engine,client *mongo.Client){
	router.Use(middleware.AuthMiddleware())

	router.GET("/movie/:imdb_id",controller.GetSingleMovieHandler(client))
	router.POST("/add-movie",controller.AddMovieHandler(client))
	router.GET("/recommended-movies",controller.GetRecommendedMoviesHandler(client))
	router.PATCH("/update-review/:imdb_id",controller.AdminReviewUpdateHandler(client))
}	