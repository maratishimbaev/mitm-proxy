package proxyInterfaces

import "mitm-proxy/app/models"

type ProxyRepository interface {
	CreateRequest(request *models.Request) (err error)
	GetRequests() (requests []models.Request, err error)
	GetRequest(id uint64) (request *models.Request, err error)
}
