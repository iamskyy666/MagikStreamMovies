package database

// Connecting to MongoDB

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DBConnect()*mongo.Client{
	err:= godotenv.Load(".env")
	if err!=nil{
		log.Println("⚠️ERROR loading .env file ---",err)
	}

	MongoDbURI:=os.Getenv("MONGODB_URI")

	if MongoDbURI == ""{
		log.Fatal("MONGODB_URI not set! ⚠️")
	}

	fmt.Println("MongoDB-URI:",MongoDbURI)

	clientOptions:=options.Client().ApplyURI(MongoDbURI)

	// Finally, connect to MongoDB
	client,err:=mongo.Connect(clientOptions)
	if err!=nil{
		log.Println("⚠️ERROR conecting to MongoDB---",err)
		return nil
	}

	return client
}

// var Client *mongo.Client = DBConnect() // Client obj.

func OpenCollection(collectionName string,client *mongo.Client)*mongo.Collection{
	err:=godotenv.Load(".env")
	if err!=nil{
		log.Println("⚠️ERROR loading .env file ---",err)
	}

	dbName:=os.Getenv("DATABASE_NAME")
	fmt.Println("DB_NAME:",dbName) // For testing purposes

	collection:= client.Database(dbName).Collection(collectionName)

	if collection==nil{
		log.Println("⚠️ ERROR loading collection---",err)
		return nil
	}

	return collection
}