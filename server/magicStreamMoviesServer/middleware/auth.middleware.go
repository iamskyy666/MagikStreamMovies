package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/utils"
)

// gin-gonic handler fx, but used in a different way
func AuthMiddleware()gin.HandlerFunc{
	return func(ctx *gin.Context){
		token,err:=utils.GetAccessToken(ctx)
		if err!=nil{
			ctx.JSON(http.StatusUnauthorized,gin.H{
				"status_code:":http.StatusUnauthorized,
			})
			ctx.Abort() //ctx.Abort() from MW's
			return 
		}
		if token==""{
				ctx.JSON(http.StatusUnauthorized,gin.H{
				"error":"⚠️ No Token Provided!",
				"status_code:":http.StatusUnauthorized,
			})
			ctx.Abort() //ctx.Abort() from MW's
			return 
		}
		claims,err:=utils.ValidateToken(token)
		if err!=nil{
			ctx.JSON(http.StatusUnauthorized,gin.H{
				"error":" ⚠️Invalid Token!",
				"status_code:":http.StatusUnauthorized,
			})
			ctx.Abort() //ctx.Abort() from MW's
			return 
		}

		ctx.Set("userId",claims.UserId)
		ctx.Set("role",claims.Role)

		ctx.Next() // Opposite of ctx.Abort()
	}
}