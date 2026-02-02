package util

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/go-sonic/sonic/util/xerr"
)

func GetClientIP(ctx interface{}) string {
	if fiberCtx, ok := ctx.(*fiber.Ctx); ok {
		return fiberCtx.IP()
	}
	return ""
}

func GetUserAgent(ctx interface{}) string {
	if fiberCtx, ok := ctx.(*fiber.Ctx); ok {
		return fiberCtx.Get("User-Agent")
	}
	return ""
}

func MustGetQueryString(ctx *fiber.Ctx, key string) (string, error) {
	str := ctx.Query(key)
	if str == "" {
		return "", xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	return str, nil
}

func MustGetQueryInt32(ctx *fiber.Ctx, key string) (int32, error) {
	str := ctx.Query(key)
	if str == "" {
		return 0, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return int32(value), nil
}

func MustGetQueryInt64(ctx *fiber.Ctx, key string) (int64, error) {
	str := ctx.Query(key)
	if str == "" {
		return 0, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func MustGetQueryInt(ctx *fiber.Ctx, key string) (int, error) {
	str := ctx.Query(key)
	if str == "" {
		return 0, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.Atoi(str)
	if err != nil {
		return 0, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func MustGetQueryBool(ctx *fiber.Ctx, key string) (bool, error) {
	str := ctx.Query(key)
	if str == "" {
		return false, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.ParseBool(str)
	if err != nil {
		return false, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func GetQueryBool(ctx *fiber.Ctx, key string, defaultValue bool) (bool, error) {
	str := ctx.Query(key)
	if str == "" {
		return defaultValue, nil
	}
	value, err := strconv.ParseBool(str)
	if err != nil {
		return false, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func ParamString(ctx *fiber.Ctx, key string) (string, error) {
	str := ctx.Params(key)
	if str == "" {
		return "", xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	return str, nil
}

func ParamInt32(ctx *fiber.Ctx, key string) (int32, error) {
	str := ctx.Params(key)
	if str == "" {
		return 0, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return int32(value), nil
}

func ParamInt64(ctx *fiber.Ctx, key string) (int64, error) {
	str := ctx.Params(key)
	if str == "" {
		return 0, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}

func ParamBool(ctx *fiber.Ctx, key string) (bool, error) {
	str := ctx.Params(key)
	if str == "" {
		return false, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.ParseBool(str)
	if err != nil {
		return false, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return value, nil
}
