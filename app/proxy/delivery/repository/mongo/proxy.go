package proxyMongo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"mitm-proxy/app/models"
)

type proxyRepository struct {
	collection *mongo.Collection
}

func NewProxyRepository(collection *mongo.Collection) *proxyRepository {
	return &proxyRepository{collection: collection}
}

func (r *proxyRepository) CreateRequest(request *models.Request) (err error) {
	_, err = r.collection.InsertOne(context.TODO(), request)

	return err
}
