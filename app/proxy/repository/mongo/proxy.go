package proxyMongo

import (
	"context"
	"github.com/kataras/golog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mitm-proxy/app/models"
)

type proxyRepository struct {
	collection *mongo.Collection
}

func NewProxyRepository(collection *mongo.Collection) *proxyRepository {
	return &proxyRepository{collection: collection}
}

func (r *proxyRepository) CreateRequest(request *models.Request) (err error) {
	golog.Info(request)

	_, err = r.collection.InsertOne(context.TODO(), request)

	return err
}

func (r *proxyRepository) GetRequests() (requests []models.Request, err error) {
	findOpts := options.Find()

	cur, err := r.collection.Find(context.TODO(), bson.D{}, findOpts)
	if err != nil {
		golog.Error(err.Error())
		return nil, err
	}

	for cur.Next(context.TODO()) {
		var request models.Request

		err := cur.Decode(&request)
		if err != nil {
			golog.Error(err.Error())
			return nil, err
		}

		requests = append(requests, request)
	}

	if err := cur.Err(); err != nil {
		golog.Error(err.Error())
		return nil, err
	}

	cur.Close(context.TODO())

	return requests, nil
}
