package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/controllers"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetUpUnProtectedRoutes(router *gin.Engine,client *mongo.Client){
   	router.GET("/movies",controller.GetMoviesHandler(client))
	router.POST("/register",controller.RegisterUserHandler(client))
	router.POST("/login",controller.LoginUserHandler(client))
	router.POST("/logout",controller.LogoutUserHandler(client))
	router.GET("/genres",controller.GetGenresHandler(client))
	router.POST("/refresh", controller.RefreshTokenHandler(client))
}	