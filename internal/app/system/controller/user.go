package controller

import (
	"context"
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/genv"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/tiger1103/gfast/v3/apiv1/system"
	commonService "github.com/tiger1103/gfast/v3/internal/app/common/service"
	"github.com/tiger1103/gfast/v3/internal/app/system/model"
	"github.com/tiger1103/gfast/v3/internal/app/system/service"
	"github.com/tiger1103/gfast/v3/library/libUtils"
)

var (
	User = UserController{}
)

type UserController struct {
	baseController
}

func (c *UserController) Login(ctx context.Context, req *system.UserLoginReq) (res *system.UserLoginRes, err error) {
	var (
		user  *model.LoginUserRes
		token string
	)
	//判断验证码是否正确
	debug := genv.GetWithCmd("gf.debug")
	if debug.Int() != 1 {
		if !commonService.Captcha().VerifyString(req.VerifyKey, req.VerifyCode) {
			err = gerror.New("验证码输入错误")
			return
		}
	}
	ip := libUtils.GetClientIp(ctx)
	userAgent := libUtils.GetUserAgent(ctx)
	user, err = service.User().GetAdminUserByUsernamePassword(ctx, req)
	if err != nil {
		// 保存登录失败的日志信息
		service.SysLoginLog().Invoke(ctx, &model.LoginLogParams{
			Status:    0,
			Username:  req.Username,
			Ip:        ip,
			UserAgent: userAgent,
			Msg:       err.Error(),
			Module:    "系统后台",
		})
		return
	}
	err = service.User().UpdateLoginInfo(ctx, user.Id, ip)
	if err != nil {
		return
	}
	// 报存登录成功的日志信息
	service.SysLoginLog().Invoke(ctx, &model.LoginLogParams{
		Status:    1,
		Username:  req.Username,
		Ip:        ip,
		UserAgent: userAgent,
		Msg:       "登录成功",
		Module:    "系统后台",
	})
	token, err = service.GfToken(ctx).GenerateToken(
		ctx,
		gconv.String(user.Id)+"-"+gmd5.MustEncryptString(user.UserName)+gmd5.MustEncryptString(user.UserPassword),
		user,
	)
	if err != nil {
		return
	}
	res = &system.UserLoginRes{
		UserInfo: user,
		Token:    token,
	}
	return
}
