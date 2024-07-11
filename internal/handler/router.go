package handler

import (
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	//_ "myapp/docs"

	"myapp/internal/usecase"
)

func NewRouter(router *gin.Engine, os usecase.ScreenRecorderUseCases, au *usecase.AuthUseCases) {

	authHandlers := &AuthHandler{
		us: *au,
	}
	screenRecorderHandlers := &ScreenRecorderHandler{
		us: os,
	}

	// Routers
	auth := router.Group("/sign-in")
	{
		auth.POST("/", authHandlers.AuthUser)
	}

	cashe := router.Group("/DeleteCasheFiles")
	{
		cashe.DELETE("/", screenRecorderHandlers.DeleteCashe)
	}

	record := router.Group("/record")
	{
		record.POST("/add", screenRecorderHandlers.AddVideo)

		d := record.Group("/download")
		{
			d.GET("/:video_id", screenRecorderHandlers.DownloadVideo)
		}
	}

	projectsUpdate := router.Group("/updatePostgresProjects")
	{
		projectsUpdate.POST("/", screenRecorderHandlers.UpdatePostgresProjects)
	}

	video := router.Group("/videos")
	{
		video.POST("/list", screenRecorderHandlers.ListVideos)

		count := video.Group("/count")
		{
			count.POST("/", screenRecorderHandlers.CountVideos)
		}

	}

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
