package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/database"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/models"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Hashing f(x) 🛡️
func HashPassword(password string)(string,error){
	HashedPassword, err:=bcrypt.GenerateFromPassword([]byte(password),bcrypt.DefaultCost)
	if err!=nil{
			log.Println("Hashing ERROR:",err.Error())
			return "",err
		}
	return string(HashedPassword),nil	
}

//! 1️⃣ POST/Add/Register User
func RegisterUserHandler(client *mongo.Client)gin.HandlerFunc{
	return func(ctx *gin.Context){
		var user models.User
		err:=ctx.ShouldBindJSON(&user); 

		if err!=nil{
			ctx.JSON(http.StatusBadRequest,gin.H{
				"error":"⚠️ Invalid Input!",
				"status_code":http.StatusBadRequest,
			})
			return 
		}
			// Validator instance
			var validate = validator.New()
			if err:= validate.Struct(user); err!=nil{
			ctx.JSON(http.StatusBadRequest,gin.H{
				"error":"⚠️ Validation failed!",
				"status_code":http.StatusBadRequest,
			})
			return 
		}
		hashedPassword,err:=HashPassword(user.Password)
		if err!=nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"⚠️ Hashing Error!",
				"status_code":http.StatusInternalServerError,
			})
			return 
		}

		var ctxt,cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var userCollection *mongo.Collection = database.OpenCollection("users",client)
		count,err:=userCollection.CountDocuments(ctxt, bson.M{"email":user.Email})
		if err!=nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"⚠️ ERROR checking existing user!",
				"status_code":http.StatusInternalServerError,
			})
			return 
		}

		// Return if user/email_ID already exists in the DB
		if count>0{
			ctx.JSON(http.StatusConflict,gin.H{
				"error":"⚠️ User with this Email-ID already exists!",
				"status_code":http.StatusConflict,
			})
			return 
		}

		user.UserID = bson.NewObjectID().Hex()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.Password = hashedPassword
		
		// finally add/register the user
		result,err:= userCollection.InsertOne(ctxt,user)

		if err!=nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"⚠️ ERROR adding/registering user!",
				"err. info":err,
				"status_code":http.StatusInternalServerError,
			})
			return 
		}
		ctx.JSON(http.StatusCreated, result)
	}
}

//! 2️⃣ POST/Log-In User
func LoginUserHandler(client *mongo.Client)gin.HandlerFunc{
	return func(ctx *gin.Context){

		var userLogin models.UserLogin

		err:=ctx.ShouldBindJSON(&userLogin);
		if err!=nil{
			ctx.JSON(http.StatusBadRequest,gin.H{
				"error":"⚠️ Invalid Input!",
				"status_code":http.StatusBadRequest,
			})
			return 
		}
		
		var ctxt,cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		
		var foundUser models.User

		var userCollection *mongo.Collection = database.OpenCollection("users",client)
		err=userCollection.FindOne(ctxt, bson.M{"email":userLogin.Email}).Decode(&foundUser)
		if err!=nil{
			ctx.JSON(http.StatusUnauthorized,gin.H{
				"error":"⚠️ Invalid email or password!",
				"status_code":http.StatusUnauthorized,
			})
			return 
		}

		// compare the entered password with the hashed-password from the DB
		err =bcrypt.CompareHashAndPassword([]byte(foundUser.Password),[]byte(userLogin.Password))
		if err!=nil{
			ctx.JSON(http.StatusUnauthorized,gin.H{
				"error":"⚠️ Invalid password!",
				"status_code":http.StatusUnauthorized,
			})
			return 
		}

		// If all ok, generate access-token 🔐
		token, refreshToken, err:= utils.GenerateAllTokens(foundUser.Email, foundUser.FirstName, foundUser.LastName, foundUser.Role, foundUser.UserID)
		if err!=nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"⚠️ Failed to GENERATE tokens!",
				"status_code":http.StatusInternalServerError,
			})
			return 
		}

		err=utils.UpdateAllTokens(foundUser.UserID, token, refreshToken,client)
		if err!=nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"⚠️ Failed to UPDATE tokens!",
				"status_code":http.StatusInternalServerError,
			})
			return 
		}

		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:  "token",
			Value: token,
			Path:  "/",
			// Domain:   "localhost",
			MaxAge:   86400,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})

		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:  "refresh_token",
			Value: refreshToken,
			Path:  "/",
			// Domain:   "localhost",
			MaxAge:   604800,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})

		// return user-resp. // Later saving them
		ctx.JSON(http.StatusOK, models.UserResponse{
			UserID: foundUser.UserID,
			FirstName: foundUser.FirstName,
			LastName: foundUser.LastName,
			Email: foundUser.Email,
			Role: foundUser.Role,
			//Token: token,
			//RefreshToken: refreshToken,
			FavouriteGenres: foundUser.FavouriteGenres,
		})
	}
}


//! 3️⃣ POST/Log-Out User
 func LogoutUserHandler(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Clear the access_token cookie

		var UserLogout struct {
			UserId string `json:"user_id"`
		}

		err := c.ShouldBindJSON(&UserLogout)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		fmt.Println("User ID from Logout request:", UserLogout.UserId)

		err = utils.UpdateAllTokens(UserLogout.UserId, "", "", client) // Clear tokens in the database
		// Optionally, we can also remove the user session from the database if needed

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error logging out"})
			return
		}
		http.SetCookie(c.Writer, &http.Cookie{
			Name:  "token",
			Value: "",
			Path:  "/",
			// Domain:   "localhost",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})

		c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully ✅"})
	}
}

 func RefreshTokenHandler(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		refreshToken, err := c.Cookie("refresh_token")

		if err != nil {
			fmt.Println("error", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unable to retrieve refresh token from cookie"})
			return
		}

		claim, err := utils.ValidateRefreshToken(refreshToken)
		if err != nil || claim == nil {
			fmt.Println("error", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
			return
		}

		var userCollection *mongo.Collection = database.OpenCollection("users", client)

		var user models.User
		err = userCollection.FindOne(ctx, bson.D{{Key: "user_id", Value: claim.UserId}}).Decode(&user)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		newToken, newRefreshToken, err := utils.GenerateAllTokens(user.Email, user.FirstName, user.LastName, user.Role, user.UserID)
		if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
		return
		}
		err = utils.UpdateAllTokens(user.UserID, newToken, newRefreshToken, client)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating tokens"})
			return
		}

		c.SetCookie("token", newToken, 86400, "/", "localhost", true, true)          // expires in 24 hours
		c.SetCookie("refresh_token", newRefreshToken, 604800, "/", "localhost", true, true) //expires in 1 week

		c.JSON(http.StatusOK, gin.H{"message": "Tokens refreshed"})
	}
}