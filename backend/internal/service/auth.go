package service

import (
	"context"

	"apikey-manager/internal/model"
)

// IAuthService 认证相关业务
type IAuthService interface {
	// Login 校验账号密码并签发 JWT
	Login(ctx context.Context, req *model.LoginReq) (*model.LoginResp, error)
	// Register 公开注册新用户（固定 role=user、status=1，不签发 token；受 app.allow_register 开关控制）
	Register(ctx context.Context, req *model.RegisterReq) (*model.RegisterResp, error)
	// Profile 当前用户信息 + 权限标识
	Profile(ctx context.Context, userID uint) (*model.ProfileResp, error)
	// ChangePassword 本人修改密码
	ChangePassword(ctx context.Context, userID uint, req *model.ChangePasswordReq) error
}

var Auth IAuthService
