package main

import (
	"context"
	"github.com/kataras/golog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/http"
)

func main() {
	col := initDB()

	p := NewProxy(col)

	golog.Info("Server started")
	golog.Fatal(http.ListenAndServe(":8000", p))
}

func initDB() *mongo.Collection {
	clientOpts := options.Client().ApplyURI("mongodb://mongo:mongo@localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		golog.Fatal(err.Error())
		return nil
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		golog.Fatal(err.Error())
		return nil
	}

	return client.Database("mitm-proxy").Collection("requests")
}
