package proxyInterfaces

import "mitm-proxy/app/models"

type ProxyUseCase interface {
	CreateRequest(request *models.Request) (err error)
	GetRequests() (requests []models.Request, err error)
	GetRequest(id uint64) (request *models.Request, err error)
	AddXXEEntity(body string) string
}
