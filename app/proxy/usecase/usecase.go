package proxyUsecase

import (
	"mitm-proxy/app/models"
	proxyInterfaces "mitm-proxy/app/proxy/interfaces"
	"sync/atomic"
)

type proxyUseCase struct {
	repository proxyInterfaces.ProxyRepository
	id uint64
}

func NewProxyUseCase(repository proxyInterfaces.ProxyRepository) *proxyUseCase {
	return &proxyUseCase{repository: repository}
}

func (u *proxyUseCase) CreateRequest(request *models.Request) (err error) {
	request.Id = atomic.AddUint64(&u.id, 1)

	return u.repository.CreateRequest(request)
}

func (u *proxyUseCase) GetRequests() (requests []models.Request, err error) {
	return u.repository.GetRequests()
}

func (u *proxyUseCase) GetRequest(id uint64) (request *models.Request, err error) {
	return u.repository.GetRequest(id)
}
