package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/database"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/models"
	"github.com/iamskyy666/MagikStreamMovies/server/magikStreamMoviesServer/utils"
	"github.com/joho/godotenv"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Validator instance
var validate = validator.New()

//! 1️⃣ GET All Movies
func GetMoviesHandler(client *mongo.Client)gin.HandlerFunc{
	return func(ctx *gin.Context) {
		ctxt,cancel:=context.WithTimeout(ctx, 100*time.Second)
		defer cancel()

		var movies []models.Movie
		var movieCollection *mongo.Collection = database.OpenCollection("movies",client)

		cursor,err:= movieCollection.Find(ctxt, bson.M{})

		if err!=nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"⚠️ Failed to fetch movies!",
				"status":http.StatusInternalServerError,
			})
			return 
		}
		defer cursor.Close(ctxt)	

		if err = cursor.All(ctxt, &movies); err!=nil{
				ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"⚠️ Failed to decode movies!",
				"status":http.StatusInternalServerError,
			})
			return 
		}
		ctx.JSON(http.StatusOK,movies)
	}
}

//! 2️⃣ GET Single Movie
func GetSingleMovieHandler(client *mongo.Client)gin.HandlerFunc{
	return func(ctx *gin.Context) {
		c,cancel:=context.WithTimeout(ctx, 100*time.Second)
		defer cancel() // always defer to free-up resources

		movieID:=ctx.Param("imdb_id") // unique-identifier
		if movieID == ""{
				ctx.JSON(http.StatusBadRequest,gin.H{
				"error":"⚠️ Movie-ID is required!",
				"status":http.StatusBadRequest,
			})
			return
		}
		var movieCollection *mongo.Collection = database.OpenCollection("movies",client)
		var movie models.Movie

		err:= movieCollection.FindOne(c, bson.M{"imdb_id":movieID}).Decode(&movie)	
		if err!=nil{
			ctx.JSON(http.StatusNotFound,gin.H{
				"error":"⚠️ Movie Not Found!",
				"status":http.StatusNotFound,
			})
			return 
		}
		ctx.JSON(http.StatusOK,movie)
	}	
}

//! 3️⃣ POST/Add Movie
func AddMovieHandler(client *mongo.Client)gin.HandlerFunc{
	return func(ctx *gin.Context) {
		c,cancel:=context.WithTimeout(ctx, 100*time.Second)
		defer cancel() // always defer to free-up resources

		var movie models.Movie
		
		err:=ctx.ShouldBindJSON(&movie); 

		if err!=nil{
			ctx.JSON(http.StatusBadRequest,gin.H{
				"error":"⚠️ Invalid Input!",
				"status_code":http.StatusBadRequest,
			})
			return 
		}

		if err:= validate.Struct(movie); err!=nil{
			ctx.JSON(http.StatusBadRequest,gin.H{
				"error":"⚠️ Validation failed!",
				"status_code":http.StatusBadRequest,
			})
			return 
		}

		// Finally, add/insert data into MongoDB
		var movieCollection *mongo.Collection = database.OpenCollection("movies",client)
		result,err:=movieCollection.InsertOne(c,movie)
		if err!=nil{
			ctx.JSON(http.StatusInternalServerError,gin.H{
				"error":"⚠️ ERROR adding movie!",
				"status_code":http.StatusInternalServerError,
			})
			return 
		}
		ctx.JSON(http.StatusCreated,result) // DONE ✅
	}
}


//! 4️⃣ Update/PATCH Admin-Review (LangChain AI 🤖🧠)
func AdminReviewUpdateHandler(client *mongo.Client)gin.HandlerFunc{
	return func(ctx *gin.Context) {

		role,err:=utils.GetRoleFromCtx(ctx)
		if err!=nil{
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":"⚠️ Member-ROLE not found in context!",
				"status_code:":http.StatusBadRequest,
			})
			return
		}

		if role!="ADMIN"{
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"error":"⚠️ User must be part of the ADMIN role!",
				"status_code:":http.StatusUnauthorized,
			})
			return
		}


		movieId:=ctx.Param("imdb_id")

			if movieId == ""{
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error":"⚠️ Movie-ID required",
				"status_code:":http.StatusBadRequest,
			})
			return 
			}

			// 2 local structs
			var req struct{
				AdminReview string `json:"admin_review"`
			}

			var resp struct{
				RankingName string `json:"ranking_name"`
				AdminReview string `json:"admin_review"`
			}

			// Bind passed-in body
			if err:=ctx.ShouldBindJSON(&req);err!=nil{
				ctx.JSON(http.StatusBadRequest, gin.H{
				"error":"⚠️ Invalid request-body!",
				"status_code:":http.StatusBadRequest,
			})
			return 
			}

			// AI to extract the sentiment of the admin-review ✨
			sentiment,rankVal,err:= GetReviewRanking(req.AdminReview,client,ctx)
			if err!=nil{
				ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":"⚠️ ERROR getting review ranking!",
				"status_code:":http.StatusInternalServerError,
			})
			return 
			}

			// Update/PATCH - 
			filter:=bson.M{"imdb_id":movieId}
			update:=bson.M{
				"$set":bson.M{
					"admin_review":req.AdminReview,
					"ranking":bson.M{
						"ranking_value":rankVal,
						"ranking_name":sentiment,
					},
				},
			}

		// clearing resources
		var ctxt,cancel = context.WithTimeout(ctx,100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies",client)
		result,err:=movieCollection.UpdateOne(ctxt,filter,update)
		if err!=nil{
				ctx.JSON(http.StatusInternalServerError, gin.H{
				"error":"⚠️ Failed to UPDATE movie!",
				"status_code:":http.StatusInternalServerError,
			})
			return 
			}

		if result.MatchedCount==0{
			ctx.JSON(http.StatusNotFound ,gin.H{
				"error":"⚠️ Movie NOT FOUND!",
				"status_code:":http.StatusNotFound,
			})
			return
		}	

		// Create a response
		resp.RankingName = sentiment
		resp.AdminReview = req.AdminReview

		ctx.JSON(http.StatusOK, resp)

		}
}

 func GetReviewRanking(admin_review string,client *mongo.Client,ctx *gin.Context) (string, int, error) {

	rankings, err := GetRankings(client,ctx)
	if err != nil {
		return "", 0, err
	}

	sentimentDelimited := ""
	for _, ranking := range rankings {
		if ranking.RankingValue != 999 {
			sentimentDelimited += ranking.RankingName + ","
		}
	}
	sentimentDelimited = strings.Trim(sentimentDelimited, ",")

	_ = godotenv.Load(".env")

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", 0, errors.New("could not read OPENAI_API_KEY")
	}

	llm, err := openai.New(openai.WithToken(apiKey))
	if err != nil {
		return "", 0, err
	}

	basePrompt := os.Getenv("BASE_PROMPT_TEMPLATE")
	prompt := strings.Replace(basePrompt, "{rankings}", sentimentDelimited, 1)

	response, err := llm.Call(ctx.Request.Context(), prompt+admin_review)
	if err != nil {
		return "", 0, err
	}

	response = strings.TrimSpace(response)

	rankVal := 999
	for _, ranking := range rankings {
		if ranking.RankingName == response {
			response = strings.TrimSpace(response) // FIX
			rankVal = ranking.RankingValue
			break
		}
	}

	return response, rankVal, nil
}


 func GetRankings(client *mongo.Client, ctx *gin.Context)([]models.Ranking,error){
	var rankings []models.Ranking

	var ctxt,cancel = context.WithTimeout(ctx,100*time.Second)
	defer cancel()

	var rankingCollection *mongo.Collection = database.OpenCollection("rankings",client)
	cursor,err:=rankingCollection.Find(ctxt, bson.M{})
	if err!=nil{
		log.Println("⚠️ ERROR:",err.Error())
		return nil,err
	}
	defer cursor.Close(ctxt)

	if err:=cursor.All(ctxt, &rankings); err!=nil{
		log.Println("⚠️ ERROR:",err.Error())
		return nil,err
	}

	return  rankings,nil
}


 //! 5️⃣ GET Recommended-Movies
 func GetRecommendedMoviesHandler(client *mongo.Client)gin.HandlerFunc{
	return func(ctx *gin.Context) {
		userId,err:=utils.GetUserIdFromCtx(ctx)
		if err!=nil{
			ctx.JSON(http.StatusBadRequest ,gin.H{
				"error":"⚠️ ID NOT FOUND IN CONTEXT!",
				"status_code:":http.StatusBadRequest,
			})
			return
		}

		// Query the db.
		favourite_genres,err:=GetUsersFavGenres(userId,client,ctx)
		if err!=nil{
			ctx.JSON(http.StatusInternalServerError ,gin.H{
				"error:":err.Error(),
				"status_code:":http.StatusInternalServerError,
			})
			return
		}

		err = godotenv.Load(".env")
		if err!=nil{
			log.Println("⚠️ WARNING: .env file not found!")
		}

		var recommendedMovieLimitVal int64 = 5
		recommendedMovieLimitStr :=os.Getenv("RECOMMENDED_MOVIE_LIMIT")

		if recommendedMovieLimitStr!=""{
			recommendedMovieLimitVal,_= strconv.ParseInt(recommendedMovieLimitStr,10,64)
		}

		findOptions:=options.Find()
		findOptions.SetSort(bson.D{{Key:"ranking.ranking_value", Value:1}})

		findOptions.SetLimit(recommendedMovieLimitVal)

		filter:=bson.M{"genre.genre_name":bson.M{"$in":favourite_genres}}

		var ctxt,cancel = context.WithTimeout(ctx,100*time.Second)
		defer cancel()

		var movieCollection *mongo.Collection = database.OpenCollection("movies",client)
		cursor,err:=movieCollection.Find(ctxt,filter,findOptions)

			

		if err!=nil{
			ctx.JSON(http.StatusInternalServerError ,gin.H{
				"error":"⚠️ ERROR fetching recommended-movies!",
				"status_code:":http.StatusInternalServerError,
			})
			return
		}

		defer cursor.Close(ctxt)
		var recommendedMovies []models.Movie

		if err:=cursor.All(ctxt,&recommendedMovies);err!=nil{
			ctx.JSON(http.StatusInternalServerError ,gin.H{
				"error":err.Error(),
				"status_code:":http.StatusInternalServerError,
			})
			return
		}

		ctx.JSON(http.StatusOK, recommendedMovies)
	}
 }


 func GetUsersFavGenres(userId string,client *mongo.Client, c *gin.Context)([]string,error){
	
	var ctxt,cancel = context.WithTimeout(c,100*time.Second)
	defer cancel()

	filter:=bson.M{"user_id":userId}
	projection:=bson.M{
		"favourite_genres.genre_name":1,
		"_id":0,
	}

	opts:= options.FindOne().SetProjection(projection)

	var results bson.M
   var userCollection *mongo.Collection = database.OpenCollection("users",client)
	err:=userCollection.FindOne(ctxt,filter,opts).Decode(&results)
	if err!=nil{
		if err==mongo.ErrNoDocuments{
			return []string{},nil
		}
		return []string{}, err // FIX: stop execution
	}
	
		favGenresArr,ok:=results["favourite_genres"].(bson.A)
		if !ok{
			return []string{},errors.New("Unable to retrieve favourite genres for user!")
		}
	
		var genreNames []string

		for _,item:=range favGenresArr{
			if genreMap, ok:=item.(bson.D);ok{
				for _,elem:=range genreMap{
					if elem.Key=="genre_name"{
						if name,ok:=elem.Value.(string);ok{
							genreNames = append(genreNames, name)
						}
					}
				}
			}
		}

		return genreNames,nil
 }

  //! 6️⃣ GET Genres
  func GetGenresHandler(client *mongo.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(c, 100*time.Second)
		defer cancel()

		var genreCollection *mongo.Collection = database.OpenCollection("genres", client)

		cursor, err := genreCollection.Find(ctx, bson.D{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching movie genres"})
			return
		}
		defer cursor.Close(ctx)

		var genres []models.Genre
		if err := cursor.All(ctx, &genres); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, genres)

	}
}
 