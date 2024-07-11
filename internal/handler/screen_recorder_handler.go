package handler

import (
	"fmt"
	"myapp/internal/models"
	"myapp/internal/usecase"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/gin-gonic/gin"
)

type ScreenRecorderHandler struct {
	us usecase.ScreenRecorderUseCases
}

// UpdatePostgresProjects перенести проекты в postgres из oracle.
//
//	@Summary		Перенести проекты из oracle в postgres
//	@Description	Добавить проекты из oracle в postgres .
//	@Tags			Update Projects In Postgres
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Router			/updatePostgresProjects/ [post]
func (h *ScreenRecorderHandler) UpdatePostgresProjects(c *gin.Context) {
	err := h.us.UpdatePostgresProjects(c.Request.Context())
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"result":  false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{
		"result":  true,
		"message": "",
		"data":    nil,
	})
}

// AddVideo добавить видео.
//
//	@Summary		Добавить видео
//	@Description	Добавить видео.
//	@Tags			Video
//	@Produce		json
//	@Param			video	body	handler.AddVideo.AddVideoRequest	true	"Добавить видео"
//	@Security		ApiKeyAuth
//	@Router			/record/add/ [post]
func (h *ScreenRecorderHandler) AddVideo(c *gin.Context) {

	type AddVideoRequest struct {
		Name        string `form:"name" binding:"required"`
		Login       string `form:"login" binding:"required"`
		SessionId   string `form:"session_id"`
		Fullpath    string `form:"fullpath" binding:"required"`
		MacAddr     string `form:"mac_addr" binding:"required"`
		IpAddr      string `form:"ip_addr"`
		GetArgs     string `form:"get_args" binding:"required"`
		ProjectUuid string `form:"project_uuid" binding:"required"`
	}

	request := AddVideoRequest{}
	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"result":  false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	video := models.Video{
		Name:        request.Name,
		Login:       request.Login,
		SessionId:   request.SessionId,
		CreatedAt:   time.Now(),
		Fullpath:    request.Fullpath,
		MacAddr:     request.MacAddr,
		IpAddr:      request.IpAddr,
		GetArgs:     request.GetArgs,
		ProjectUUID: request.ProjectUuid,
	}

	err := h.us.AddVideo(c.Request.Context(), video)
	if err != nil {
		log.Error().Err(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"result":  false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"result":  true,
		"message": "Видео добавлено в базу",
		"data":    nil,
	})
}

// DownloadVideo Выгрузить видео
//
//	@Summary		Выгрузить видео
//	@Description	Выгрузить видео.
//	@Tags			Video
//	@Produce		json
//	@Param			id	path	string	true	"ID видео"
//	@Success		200	{file}	file
//	@Security		ApiKeyAuth
//	@Router			/record/download/{id}/ [get]
func (h *ScreenRecorderHandler) DownloadVideo(c *gin.Context) {

	var directory string = "./tmp/resultVideos"

	video_id := c.Param("video_id")

	fileName, info, err := h.us.DownloadVideo(c.Request.Context(), video_id)
	if err != nil {
		log.Error().Err(err)
		switch {
		case strings.Contains(err.Error(), "Видео не существует"):
			c.JSON(http.StatusNotFound, gin.H{
				"result":  false,
				"message": err.Error(),
				"data":    nil,
			})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"result":  false,
				"message": err.Error(),
				"data":    nil,
			})
			return
		}
	}

	fullPath := fmt.Sprintf("%s/%s", directory, fileName)

	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"message": info,
		"data":    nil,
	})

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")
	c.File(fullPath)
}

// DeleteCashe очистить кэш.
//
//	@Summary		Удалить временные файлы
//	@Description	Удалить временные файлы.
//	@Tags			Cashe
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Router			/DeleteCasheFiles/ [delete]
func (h *ScreenRecorderHandler) DeleteCashe(c *gin.Context) {
	err := h.us.DeleteCashe(c.Request.Context())
	if err != nil {
		log.Error().Err(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"result":  false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"result":  true,
		"message": "",
		"data":    nil,
	})
}

// ListVideos получить список всех видео.
//
//	@Summary		Получить список всех видео
//	@Description	Получить список всех видео.
//	@Tags			Video
//	@Produce		json
//	@Param			query	body	models.ListVideosRequest	true	"Данные для поиска"
//	@Success		200		{array}	[]models.ListVideos
//	@Security		ApiKeyAuth
//	@Router			/videos/list/ [get]
func (h *ScreenRecorderHandler) ListVideos(c *gin.Context) {

	request := models.ListVideosRequest{}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"result":  false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	//log.Info().Msgf("handler, ListVideos, login %s", request.SearchValue, request.SearchValue, request.Offset)

	videos, err := h.us.ListVideos(c.Request.Context(), request.SearchValue, request.OrderValue, request.Offset)
	if err != nil {
		log.Error().Err(err)

		c.JSON(http.StatusInternalServerError, gin.H{
			"result":  false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"message": "",
		"data":    videos,
	})
}

// ListVideosOffset Посчитать количество видео.
//
//	@Summary		Количество видео
//	@Description	Количество видео.
//	@Tags			Video
//	@Produce		json
//	@Param			query	body	models.ListVideosRequest	true	"Данные для поиска"
//	@Success		200		integer
//	@Security		ApiKeyAuth
//	@Router			/countVideos/ [get]
func (h *ScreenRecorderHandler) CountVideos(c *gin.Context) {

	request := models.ListVideosRequest{}

	if err := c.ShouldBind(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"result":  false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	videos, err := h.us.CountVideos(c.Request.Context(), request.SearchValue, request.OrderValue, request.Offset)
	if err != nil {
		log.Error().Err(err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"result":  false,
			"message": err.Error(),
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"result":  true,
		"message": "",
		"data":    videos,
	})
}
