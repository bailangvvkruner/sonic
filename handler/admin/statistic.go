package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/service"
)

type StatisticHandler struct {
	StatisticService service.StatisticService
}

func NewStatisticHandler(l service.StatisticService) *StatisticHandler {
	return &StatisticHandler{
		StatisticService: l,
	}
}

func (s *StatisticHandler) Statistics(ctx *fiber.Ctx) (interface{}, error) {
	return s.StatisticService.Statistic(ctx.UserContext())
}

func (s *StatisticHandler) StatisticsWithUser(ctx *fiber.Ctx) (interface{}, error) {
	return s.StatisticService.StatisticWithUser(ctx.UserContext())
}

