package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/poke-factory/cheri-berry/internal/handlers"
	"github.com/poke-factory/cheri-berry/internal/middlewares"
	// 导入其他需要的包
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.GET(":package", handlers.GetPackageInfo)
	r.GET(":package/-/:file", handlers.GetPackageFile)

	prefix := r.Group("-")
	prefix.PUT("/user/:id", handlers.Login)
	prefix.DELETE("/user/token/:accessToken", handlers.Logout)

	r.Use(middlewares.AuthMiddleware())
	r.PUT(":package", handlers.UploadPackage)

	return r
}
