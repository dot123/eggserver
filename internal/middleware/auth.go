package middleware

import (
	"eggServer/internal/config"
	"eggServer/internal/constant"
	"eggServer/internal/contextx"
	"eggServer/internal/ginx"
	"eggServer/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"net/http"
	"strings"
)

// Auth 是一个认证中间件，用于验证传入请求中的 JWT token
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取 Authorization 字段，格式应为 "Bearer <token>"
		tokenHeader := c.GetHeader("Authorization")
		if tokenHeader == "" {
			// 如果 Authorization 头部不存在，则返回 401 未授权错误，并终止请求处理
			ginx.ResError(c, http.StatusUnauthorized, errors.NewResponseError(constant.TokenInvalid, nil))
			c.Abort()
			return
		}

		// 将 Authorization 头部按照空格分割，以获取 token 部分
		parts := strings.Split(tokenHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			// 如果格式不正确（必须是 "Bearer <token>"），返回 401 未授权错误，并终止请求处理
			ginx.ResError(c, http.StatusUnauthorized, errors.NewResponseError(constant.TokenInvalid, nil))
			c.Abort()
			return
		}

		// 提取 token 部分
		tokenString := parts[1]
		claims := jwt.MapClaims{}

		// 解析 token，获取其声称的内容
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// 提供签名验证所需的密钥
			return []byte(config.C.JWTAuth.Key), nil
		})

		// 登录已过期
		if err != nil && errors.Is(err, jwt.ErrTokenExpired) {
			ginx.ResError(c, http.StatusUnauthorized, errors.NewResponseError(constant.LoginExpired, nil))
			c.Abort()
			return
		}

		if err != nil || !token.Valid {
			// 如果 token 解析失败或无效，返回 401 未授权错误，并终止请求处理
			ginx.ResError(c, http.StatusUnauthorized, errors.NewResponseError(constant.TokenInvalid, nil))
			c.Abort()
			return
		}

		// 检查 token 是否包含 userId
		if claims["userId"] == nil {
			// 如果 token 中缺少 userId，返回 401 未授权错误，并终止请求处理
			ginx.ResError(c, http.StatusUnauthorized, errors.NewResponseError(constant.Unauthorized, nil))
			c.Abort()
			return
		}

		if config.C.JWTAuth.UseSession {
			// 检查 token 是否包含 sessionId
			if claims["sessionId"] == nil {
				// 如果 token 中缺少 sessionId，返回 401 未授权错误，并终止请求处理
				ginx.ResError(c, http.StatusUnauthorized, errors.NewResponseError(constant.Unauthorized, nil))
				c.Abort()
				return
			}
		}

		// 从上下文中获取 Redis 客户端实例
		ctx := c.Request.Context()
		rb := contextx.FromRB(ctx)

		if config.C.JWTAuth.UseSession {
			// 从 Redis 中获取当前用户的 sessionId
			sessionId, err := rb.Client().Get(ctx, cast.ToString(claims["userId"])).Result()
			if err != nil {
				// 如果获取 session 失败（例如，session 已过期），返回 401 登录过期错误，并终止请求处理
				ginx.ResError(c, http.StatusUnauthorized, errors.NewResponseError(constant.LoginExpired, nil))
				c.Abort()
				return
			}

			// 验证 sessionId 是否与请求中的 sessionId 匹配
			if sessionId != cast.ToString(claims["sessionId"]) {
				// 如果 token 与 Redis 中的 session 不匹配（例如，用户在其他设备上登录），返回 401 已在其他设备登录错误，并终止请求处理
				ginx.ResError(c, http.StatusUnauthorized, errors.NewResponseError(constant.LoggedInOnAnotherDevice, nil))
				c.Abort()
				return
			}
		}

		// 将用户信息（userId 和 roleId）存入请求上下文中，以便后续处理器可以使用
		ctx = contextx.NewUserID(ctx, cast.ToUint64(claims["userId"]))

		if claims["roleId"] != nil {
			ctx = contextx.NewRoleID(ctx, cast.ToUint64(claims["roleId"]))
		}
		c.Request = c.Request.WithContext(ctx)

		// 继续处理请求
		c.Next()
	}
}
