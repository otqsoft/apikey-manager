package logic

import (
	"context"
	"regexp"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"golang.org/x/crypto/bcrypt"

	"apikey-manager/internal/model"
	"apikey-manager/internal/pkg/errs"
	"apikey-manager/internal/pkg/jwtx"
	"apikey-manager/internal/service"
)

// usernameRe 注册用户名格式：3~32 位字母 / 数字 / 下划线（契约 8.1）
var usernameRe = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)

func init() {
	service.Auth = &authService{}
}

type authService struct{}

// Login 账号密码登录，成功签发 JWT
func (s *authService) Login(ctx context.Context, req *model.LoginReq) (*model.LoginResp, error) {
	var u model.User
	err := g.DB().Model("users").Where("username", req.Username).Scan(&u)
	if err != nil {
		g.Log().Errorf(ctx, "登录查询用户失败: %v", err)
		return nil, errs.BadRequest("账号或密码错误")
	}
	if u.ID == 0 {
		return nil, errs.BadRequest("账号或密码错误")
	}
	if u.Status != 1 {
		return nil, errs.Forbidden("账号已被停用，请联系管理员")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errs.BadRequest("账号或密码错误")
	}
	token, err := jwtx.Issue(u.ID, u.Username, u.Role)
	if err != nil {
		g.Log().Errorf(ctx, "签发 token 失败: %v", err)
		return nil, errs.BadRequest("登录失败，请稍后重试")
	}
	resp := &model.LoginResp{}
	resp.Token = token
	resp.User.ID = u.ID
	resp.User.Username = u.Username
	resp.User.Role = u.Role
	resp.User.DisplayName = u.DisplayName
	return resp, nil
}

// Register 公开注册新用户（契约 8.1）：开关控制、用户名唯一、bcrypt 哈希；固定 role=user、status=1；不签发 token
func (s *authService) Register(ctx context.Context, req *model.RegisterReq) (*model.RegisterResp, error) {
	// 注册总开关：配置缺失按 true 兜底
	if !g.Cfg().MustGet(ctx, "app.allow_register", true).Bool() {
		return nil, errs.BadRequest("注册已关闭")
	}
	// 格式二次兜底校验（与 DTO v tag 双保险）
	if !usernameRe.MatchString(req.Username) {
		return nil, errs.BadRequest("用户名须为3~32位字母、数字或下划线")
	}
	if len(req.Password) < 6 {
		return nil, errs.BadRequest("密码至少6位")
	}
	// 用户名唯一
	cnt, err := g.DB().Model("users").Where("username", req.Username).Count()
	if err != nil {
		g.Log().Errorf(ctx, "注册查询用户名失败: %v", err)
		return nil, errs.BadRequest("注册失败，请稍后重试")
	}
	if cnt > 0 {
		return nil, errs.BadRequest("用户名已存在")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		g.Log().Errorf(ctx, "注册密码加密失败: %v", err)
		return nil, errs.BadRequest("注册失败，请稍后重试")
	}
	displayName := strings.TrimSpace(req.DisplayName)
	if displayName == "" {
		displayName = req.Username // 昵称为空回退用户名
	}
	id, err := g.DB().Model("users").InsertAndGetId(g.Map{
		"username":      req.Username,
		"password_hash": string(hash),
		"role":          "user", // 后端固定普通用户，禁止接收 role
		"display_name":  displayName,
		"status":        1, // 后端固定启用
	})
	if err != nil {
		g.Log().Errorf(ctx, "注册写入用户失败: %v", err)
		return nil, errs.BadRequest("注册失败，请稍后重试")
	}
	return &model.RegisterResp{
		ID:          uint(id),
		Username:    req.Username,
		Role:        "user",
		DisplayName: displayName,
	}, nil
}

// Profile 当前用户信息 + 权限标识
func (s *authService) Profile(ctx context.Context, userID uint) (*model.ProfileResp, error) {
	var u model.User
	err := g.DB().Model("users").Where("id", userID).Scan(&u)
	if err != nil || u.ID == 0 {
		return nil, errs.Unauthorized("用户不存在或已被删除")
	}
	resp := &model.ProfileResp{
		ID:          u.ID,
		Username:    u.Username,
		Role:        u.Role,
		DisplayName: u.DisplayName,
		Status:      u.Status,
	}
	if u.Role == "admin" {
		resp.Permissions = []string{"*"}
	} else {
		resp.Permissions = []string{"chat:use"}
	}
	return resp, nil
}

// ChangePassword 本人修改密码
func (s *authService) ChangePassword(ctx context.Context, userID uint, req *model.ChangePasswordReq) error {
	var u model.User
	if err := g.DB().Model("users").Where("id", userID).Scan(&u); err != nil || u.ID == 0 {
		return errs.Unauthorized("用户不存在")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.OldPassword)); err != nil {
		return errs.BadRequest("旧密码不正确")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errs.BadRequest("密码加密失败")
	}
	_, err = g.DB().Model("users").Where("id", userID).Update(g.Map{
		"password_hash": string(hash),
	})
	return err
}
