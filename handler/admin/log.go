package admin

import (
	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/model/dto"
	"github.com/go-sonic/sonic/model/param"
	"github.com/go-sonic/sonic/service"
	"github.com/go-sonic/sonic/util"
	"github.com/go-sonic/sonic/util/xerr"
)

type LogHandler struct {
	LogService service.LogService
}

func NewLogHandler(logService service.LogService) *LogHandler {
	return &LogHandler{
		LogService: logService,
	}
}

func (l *LogHandler) PageLatestLog(ctx *fiber.Ctx) (interface{}, error) {
	top, err := util.MustGetQueryInt32(ctx, "top")
	if err != nil {
		top = 10
	}
	logs, _, err := l.LogService.PageLog(ctx.UserContext(), param.Pagination{PageSize: int(top)}, &param.Sort{Fields: []string{"createTime,desc"}})
	if err != nil {
		return nil, err
	}
	logDTOs := make([]*dto.Log, 0, len(logs))
	for _, log := range logs {
		logDTOs = append(logDTOs, l.LogService.ConvertToDTO(log))
	}
	return logDTOs, nil
}

func (l *LogHandler) PageLog(ctx *fiber.Ctx) (interface{}, error) {
	type LogParam struct {
		param.Pagination
		*param.Sort
	}
	var logParam LogParam
	err := ctx.QueryParser(&logParam)
	if err != nil {
		return nil, xerr.WithMsg(err, "parameter error").WithStatus(xerr.StatusBadRequest)
	}
	if logParam.Sort == nil {
		logParam.Sort = &param.Sort{
			Fields: []string{"createTime,desc"},
		}
	}
	logs, totalCount, err := l.LogService.PageLog(ctx.UserContext(), logparam.Pagination, logParam.Sort)
	if err != nil {
		return nil, err
	}
	logDTOs := make([]*dto.Log, 0, len(logs))
	for _, log := range logs {
		logDTOs = append(logDTOs, l.LogService.ConvertToDTO(log))
	}
	return dto.NewPage(logDTOs, totalCount, logparam.Pagination), nil
}

func (l *LogHandler) ClearLog(ctx *fiber.Ctx) (interface{}, error) {
	return nil, l.LogService.Clear(ctx.UserContext())
}

