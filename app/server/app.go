package server

import (
	"context"
	"github.com/kataras/golog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	proxyHttp "mitm-proxy/app/proxy/delivery/http"
	proxyInterfaces "mitm-proxy/app/proxy/delivery/interfaces"
	proxyMongo "mitm-proxy/app/proxy/delivery/repository/mongo"
	proxyUsecase "mitm-proxy/app/proxy/delivery/usecase"
	"net/http"
)

type app struct {
	proxyUseCase proxyInterfaces.ProxyUseCase
}

func NewApp() *app {
	collection := initDatabase()

	proxyRepository := proxyMongo.NewProxyRepository(collection)

	return &app{
		proxyUseCase: proxyUsecase.NewProxyUseCase(proxyRepository),
	}
}

func (a *app) Start() {
	h := proxyHttp.NewHandler(a.proxyUseCase)

	golog.Info("Server started")
	golog.Fatal(http.ListenAndServe(":8000", h))
}

func initDatabase() *mongo.Collection {
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
