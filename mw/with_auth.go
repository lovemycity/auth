package mw

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

func WithAuth(ctx *gin.Context) {
	s := sessions.Default(ctx)
	if u := s.Get("@user"); u != nil {
		ctx.Set("@user", u)
		ctx.Next()
	} else {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		ctx.Abort()
	}
}
