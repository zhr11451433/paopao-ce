package middleware

import (
	"net/http"
	"paopao/internal/auth"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//c.GetHeader("Authorization")：从HTTP 请求头里取值。
		//c.Get("user_id")：从 Gin 的上下文（context） 里取值，通常是之前用 c.Set("user_id", ...) 存进去的。
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		tokenString := parts[1]
		claims := &auth.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 无效或已过期"})
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Next()
	}
}

func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, _ := c.Get("role")
		switch v {
		case "admin":
			c.Next()
		case "customer":
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		default:
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "权限不足"})
			return
		}
	}
}


// OptionalAuth：和 Auth 一样解析 Bearer token，但【没 token / token 无效时一律放行】，
// 只是不往上下文里塞 user_id。
//
// 【为什么需要它】
//   GET /posts（广场列表）、GET /posts/:id（帖子详情）是公开路由，游客也要能看，
//   但它们要返回 is_liked —— 这就需要知道"当前是谁"。
//   Auth 中间件遇到没带 token 的请求会直接 401，把游客挡在门外，所以不能直接用。
//
// 【为什么 token 过期也不报错】
//   游客能看的接口，不该因为"登录态过期了"就整个页面炸掉。
//   过期就当游客处理（is_liked 恒 false），用户去点需要登录的按钮时才会被拦。
//
// 【用法】挂在公开路由上，handler 里用 optionalUserID(c) 取值，取不到就是 0。
func OptionalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Fields(header)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}
		claims := &auth.Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(t *jwt.Token) (any, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.Next()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Next()
	}
}