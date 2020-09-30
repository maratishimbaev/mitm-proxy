package server

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/kataras/golog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	proxyHttp "mitm-proxy/app/proxy/delivery/http"
	proxyInterfaces "mitm-proxy/app/proxy/interfaces"
	proxyMongo "mitm-proxy/app/proxy/repository/mongo"
	proxyUsecase "mitm-proxy/app/proxy/usecase"
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

	go func() {
		golog.Info("Proxy server started at port 8000")
		golog.Fatal(http.ListenAndServe(":8000", h))
	}()

	router := mux.NewRouter()
	router.HandleFunc("/requests", h.GetRequests).Methods("GET")
	router.HandleFunc("/requests/{id:[0-9]+}", h.RepeatRequest).Methods("GET")
	router.HandleFunc("/requests/{id:[0-9]+}/check", h.CheckXXE).Methods("GET")

	http.Handle("/", router)

	golog.Info("Repeater server started at port 8001")
	err := http.ListenAndServe(":8001", nil)
	if err != nil {
		golog.Error("Repeater server failed: ", err.Error())
	}
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

	collection := client.Database("mitm-proxy").Collection("requests")
	_, _ = collection.DeleteMany(context.TODO(), bson.D{})
	return collection
}
