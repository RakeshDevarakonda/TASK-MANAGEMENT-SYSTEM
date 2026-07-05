package middleware

import (
    "net/http"
    "task-system/internal/utils"
    "github.com/gin-gonic/gin"
)

// ExtractTokenClaims validates the JWT access token (if present) and stores the claims in the Gin context.
// It does not enforce authorization – that is left to downstream middleware (e.g., RequireAuth).
// If the token is malformed or verification fails, the request is aborted with 401.
func ExtractTokenClaims() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr, err := c.Cookie("accessToken")
        if err != nil {
            // No token – just continue; downstream middleware can decide what to do.
            c.Next()
            return
        }
        claims, err := utils.ValidateAccessToken(tokenStr)
        if err != nil {
            utils.SendError(c, http.StatusUnauthorized, "unauthorized: invalid token", nil)
            c.Abort()
            return
        }
        // Store claims for later use.
        c.Set("tokenClaims", claims)
        c.Next()
    }
}

// TokenClaimsFromContext retrieves the *utils.TokenClaims stored by ExtractTokenClaims.
// Returns the claims and a bool indicating if they were present.
func TokenClaimsFromContext(c *gin.Context) (*utils.TokenClaims, bool) {
    if v, exists := c.Get("tokenClaims"); exists {
        if claims, ok := v.(*utils.TokenClaims); ok {
            return claims, true
        }
    }
    return nil, false
}
