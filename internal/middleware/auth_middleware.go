package middleware

import (
	"database/sql"
	"net/http"
	"time"

	"task-system/internal/models"
	"task-system/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RequireAuth is a stateful JWT verification middleware that checks PostgreSQL to validate active sessions
func RequireAuth(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr, err := c.Cookie("accessToken")
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized: missing access token", nil)
			c.Abort()
			return
		}

		claims, err := utils.ValidateAccessToken(tokenStr)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized: invalid token", nil)
			c.Abort()
			return
		}

		userID := claims.UserID
		sessionID := claims.SessionID

		// Check 1: Verify session ID in PostgreSQL exists and belongs to the user
		var session models.Session
		sessionQuery := "SELECT id, user_id, expires_at FROM sessions WHERE id = $1"
		err = db.Get(&session, sessionQuery, sessionID)
		if err == sql.ErrNoRows || session.UserID != userID {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized: session expired or active on another device", nil)
			c.Abort()
			return
		} else if err != nil {
			utils.SendError(c, http.StatusInternalServerError, "failed to check session database", nil)
			c.Abort()
			return
		}

		// Check 2: Verify the session has not expired
		if session.ExpiresAt < time.Now().Unix() {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized: session expired", nil)
			c.Abort()
			return
		}

		// Check 3: Load user details directly from PostgreSQL
		var user models.User
		userQuery := "SELECT id, name, email, created_at, updated_at FROM users WHERE id = $1"
		err = db.Get(&user, userQuery, userID)
		if err != nil {
			utils.SendError(c, http.StatusUnauthorized, "unauthorized: user not found", nil)
			c.Abort()
			return
		}

		c.Set("currentUser", &user)
		c.Next()
	}
}
