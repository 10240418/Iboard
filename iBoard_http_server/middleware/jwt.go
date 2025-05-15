package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func extractToken(authHeader string) string {
	// 检查并移除 "Bearer " 前缀
	if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
		return authHeader[7:]
	}
	return authHeader
}

func AuthorizeJWTAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "no token found",
			})
			return
		}

		// 提取 token
		tokenString := extractToken(authHeader)
		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		if token.Valid {
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "invalid token claims",
				})
				return
			}

			if claims["isAdmin"] == false {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "unauthorized",
				})
				return
			}

			// Set claims directly without conversion
			c.Set("email", claims["email"])
			c.Set("claims", claims)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}
	}
}

func AuthorizeJWTStaff() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "no token found",
			})
			return
		}

		// 提取 token
		tokenString := extractToken(authHeader)
		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": err.Error(),
			})
			return
		}

		if token.Valid {
			claims := token.Claims.(jwt.MapClaims)

			if claims["isAdmin"] == true {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"message": "unauthorized",
				})
				return
			}

			c.Set("email", claims["email"])
			c.Next()
		} else {
			fmt.Println(err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "invalid token",
			})
			return
		}
	}
}

func AuthorizeJWTBuildingAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		tokenString := extractToken(authHeader)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		if claims["isBuildingAdmin"] != true {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		c.Set("email", claims["email"])
		c.Next()
	}
}

func AuthorizeJWTBuilding() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		tokenString := extractToken(authHeader)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		if claims["isBuilding"] != true {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		// Convert jwt.MapClaims to map[string]interface{} before setting in context
		claimsMap := make(map[string]interface{})
		for key, value := range claims {
			claimsMap[key] = value
		}

		// Set claims in context for later use
		c.Set("claims", claimsMap)
		c.Next()
	}
}

// AuthorizeJWTUpload is a middleware that authorizes both superadmin and buildingadmin tokens
func AuthorizeJWTUpload() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		tokenString := extractToken(authHeader)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		// Check if the token is either for a superadmin or buildingadmin
		isAdmin, _ := claims["isAdmin"].(bool)
		isBuildingAdmin, _ := claims["isBuildingAdmin"].(bool)

		if !isAdmin && !isBuildingAdmin {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		// Set claims in context for later use
		c.Set("claims", claims)
		c.Set("email", claims["email"])
		c.Next()
	}
}

func AuthorizeJWTDevice() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		tokenString := extractToken(authHeader)
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token format",
			})
			return
		}

		token, err := base_services.NewJWTService().ValidateToken(tokenString)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token claims",
			})
			return
		}

		if claims["isDevice"] != true {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Insufficient permissions",
			})
			return
		}

		// Convert jwt.MapClaims to map[string]interface{} before setting in context
		claimsMap := make(map[string]interface{})
		for key, value := range claims {
			claimsMap[key] = value
		}

		// Set claims in context for later use
		c.Set("claims", claimsMap)
		c.Next()
	}
}
