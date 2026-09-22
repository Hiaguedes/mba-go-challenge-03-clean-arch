package usecase

import (
	"github.com/hiaguedes/mba-go-challenge-03-clean-arch/internal/entity"
)

type ListOrdersUseCase struct {
	OrderRepository entity.OrderRepositoryInterface
}

type ListOrdersOutputDTO struct {
	Orders []entity.Order `json:"orders"`
}

func NewListOrdersUseCase(orderRepository entity.OrderRepositoryInterface) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		OrderRepository: orderRepository,
	}
}

func (c *ListOrdersUseCase) Execute() (ListOrdersOutputDTO, error) {
	orders, err := c.OrderRepository.ListOrders()
	if err != nil {
		return ListOrdersOutputDTO{}, err
	}
	dto := make([]entity.Order, len(orders))
	for i, order := range orders {
		dto[i] = *order
	}
	
	return ListOrdersOutputDTO{Orders: dto}, nil
}