package utils

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Role      string
	UserId    string
	jwt.RegisteredClaims
}

var JWT_SECRET_KEY string
var JWT_REFRESH_SECRET_KEY string

// FIX: load env before using keys (in init)
func init() {
	JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
	JWT_REFRESH_SECRET_KEY = os.Getenv("JWT_REFRESH_SECRET_KEY")
}


func GenerateAllTokens(email,firstName, lastName, role, userId string)(string,string,error){

	// First, access-token
	claims:=&SignedDetails{
		Email: email,
		FirstName: firstName,
		LastName: lastName,
		Role:role,
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:"MagikStream",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24*time.Hour)),
		},
	}

	token:=jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
	signedToken,err:=token.SignedString([]byte(JWT_SECRET_KEY))


	if err!=nil{
		log.Println("⚠️ERROR:",err.Error())
		return "","",err
	}

	// Now, refresh-token
	refreshClaims:=&SignedDetails{
		Email: email,
		FirstName: firstName,
		LastName: lastName,
		Role:role,
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:"MagikStream",
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24*7*time.Hour)),
		},
	}

	refreshToken:= jwt.NewWithClaims(jwt.SigningMethodHS256,refreshClaims)
	signedRefreshToken,err:=refreshToken.SignedString([]byte(JWT_REFRESH_SECRET_KEY))

	if err!=nil{
		log.Println("⚠️ERROR:",err.Error())
		return "","",err
	}

	return signedToken, signedRefreshToken, nil
}

func UpdateAllTokens(userId,token,refreshToken string,client *mongo.Client)(err error){
	var ctx,cancel= context.WithTimeout(context.Background(),100*time.Second)
	defer cancel()

	updatedAt,_:=time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))

	updatedData:=bson.M{
		"$set": bson.M{
			"token":token,
			"refresh_token":refreshToken,
			"updated_at":updatedAt,
		},
	}

	// Open collection
   var userCollection *mongo.Collection = database.OpenCollection("users",client)
	_,err=userCollection.UpdateOne(ctx, bson.M{"user_id":userId},updatedData)
	if err!=nil{
		log.Println("⚠️ERROR:",err.Error())
		return err
	}
	return nil
}

// Get/extract access-token from the req-header. (for auth-mw)
func GetAccessToken(ctx *gin.Context)(string, error){
	// authHeader:=ctx.Request.Header.Get("Authorization")
	
	// if authHeader==""{
	// 	return "",errors.New("⚠️ Authorization-header is required!")
	// }

	// // FIX: prevent slice out-of-range panic
	// if !strings.HasPrefix(authHeader, "Bearer ") {
	// return "", errors.New("invalid authorization format")
	// }

	// tokenStr:= authHeader[len("Bearer "):] // exclude "Bearer "

	// if tokenStr==""{
	// 	return "",errors.New("⚠️ Bearer token is required!")
	// }

	tokenStr,err:=ctx.Cookie("token")
	if err!=nil{
		return  "",err
	}

	return tokenStr,nil
}

// Validate the token (for auth-mw)
func ValidateToken(tokenStr string)(*SignedDetails,error){
claims:=&SignedDetails{}

token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token)(interface{}, error){
	return []byte(JWT_SECRET_KEY), nil
})

if err != nil { // FIX: check err first
	return nil, err
}

if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok { // FIX
	return nil, errors.New("unexpected signing method")
}

if !token.Valid { // FIX
	return nil, errors.New("invalid token")
}

if claims.ExpiresAt.Time.Before(time.Now()) {
	return nil, errors.New("⚠️ Token has expired!")
}

return claims,nil
}

// for GetRecommendedMoviesHandler()
func GetUserIdFromCtx(ctx *gin.Context)(string,error){
	userId, exists:= ctx.Get("userId")

	if !exists{
		return "",errors.New("userId does not exist in this context!")
	}

	id,ok:=userId.(string)

	if !ok{
		return "",errors.New("Unable to retrieve userId!")
	}
	return id,nil
}

// for AdminReviewUpdateHandler()
func GetRoleFromCtx(ctx *gin.Context)(string,error){
	role, exists:= ctx.Get("role")

	if !exists{
		return "",errors.New("MEMBER-ROLE does not exist in this context!")
	}

	memberRole,ok:=role.(string)

	if !ok{
		return "",errors.New("Unable to retrieve member-role!")
	}
	return memberRole,nil
}

func ValidateRefreshToken(tokenString string) (*SignedDetails, error) {
	claims := &SignedDetails{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {

		return []byte(JWT_REFRESH_SECRET_KEY), nil
	})

	if err != nil {
		return nil, err
	}

	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("refresh token has expired")
	}

	return claims, nil
}