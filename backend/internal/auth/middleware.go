package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"sigefae/internal/db"
)

func Middleware(database *gorm.DB) gin.HandlerFunc {

	return func(c *gin.Context) {

		header := c.GetHeader("Authorization")

		if header == "" {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token requerido",
			})

			c.Abort()
			return
		}

		if !strings.HasPrefix(header, "Bearer ") {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token inválido",
			})

			c.Abort()
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")

		claims, err := ValidateToken(token)

		if err != nil {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "token inválido",
			})

			c.Abort()
			return
		}

		var user db.Usuario

		err = database.
			Preload("Rol").
			Preload("Cargo").
			First(&user, claims.UserID).Error

		if err != nil {

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "usuario no encontrado",
			})

			c.Abort()
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)

		if user.Rol != nil {
			c.Set("rol", user.Rol.Nombre)
		}

		c.Next()
	}
}
