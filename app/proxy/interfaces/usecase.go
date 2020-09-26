package proxyInterfaces

import "mitm-proxy/app/models"

type ProxyUseCase interface {
	CreateRequest(request *models.Request) (err error)
}
