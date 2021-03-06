package api

import (
	"github.com/duchenhao/backend-demo/internal/api/forms"
	"github.com/duchenhao/backend-demo/internal/bus"
	"github.com/duchenhao/backend-demo/internal/model"
)

func signUp(ctx *model.ReqContext, form *forms.SignUpForm) Response {
	if form.Password != form.Password2 {
		ctx.Logger.Info("password not match")
		return Error(400, "密码不匹配")
	}

	existing := &model.GetUserByNameQuery{
		Ctx:  ctx,
		Name: form.Name,
	}
	if err := bus.Dispatch(existing); err != nil && err != model.ErrUserNotFound {
		ctx.Logger.Error(err.Error())
		return ServerError()
	}
	if existing.User != nil {
		ctx.Logger.Info("user exists")
		return Error(400, "用户已存在")
	}

	cmd := &model.SignUpCommand{}
	cmd.Ctx = ctx
	cmd.Name = form.Name
	cmd.Password = form.Password
	if err := bus.Dispatch(cmd); err != nil {
		ctx.Logger.Error(err.Error())
		return ServerError()
	}

	user := cmd.User
	tokenCmd := &model.CreateTokenCommand{
		User: user,
	}
	if err := bus.Dispatch(tokenCmd); err != nil {
		ctx.Logger.Error(err.Error())
		return ServerError()
	}

	return JSON(tokenCmd.Result)
}

func login(ctx *model.ReqContext, form *forms.LoginForm) Response {
	query := &model.LoginQuery{}
	query.Ctx = ctx
	query.Name = form.Name
	query.Password = form.Password
	if err := bus.Dispatch(query); err != nil {
		ctx.Logger.Error(err.Error())
		if err == model.ErrInvalidPassword {
			return AuthError()
		}
		return ServerError()
	}

	tokenCmd := &model.CreateTokenCommand{
		User: query.User,
	}
	if err := bus.Dispatch(tokenCmd); err != nil {
		ctx.Logger.Error(err.Error())
		return ServerError()
	}

	return JSON(tokenCmd.Result)
}

func refreshToken(ctx *model.ReqContext, form *forms.RefreshTokenForm) Response {
	query := &model.RefreshTokenCommand{}
	query.TokenPair = &model.TokenPair{RefreshToken: form.Token}
	if err := bus.Dispatch(query); err != nil {
		ctx.Logger.Error(err.Error())
		return ServerError()
	}
	return JSON(query.TokenPair)
}

func getUserInfo(ctx *model.ReqContext) Response {
	ret := &model.UserInfo{
		UserId: ctx.SignedInUser.UserId,
		Stores: make([]*model.Store, 0),
	}

	return JSON(ret)
}
