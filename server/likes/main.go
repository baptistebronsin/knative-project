package main

import (
	"encoding/json"
	"likes/domain"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func likesHandler(service domain.LikeService) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			println("Upgrade error:", err)
			return err
		}
		defer conn.Close()

		println("WebSocket connection established on /likes")

		sendLikes := func() {
			likes, err := service.GetLikes()
			if err != nil {
				println("Error fetching likes:", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Failed to retrieve likes"}`))
				return
			}

			data, _ := json.Marshal(likes)
			conn.WriteMessage(websocket.TextMessage, data)
		}

		sendLikes()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sendLikes()
			}
		}
	}
}

func getLikes(service domain.LikeService) echo.HandlerFunc {
	return func(c echo.Context) error {
		likes, err := service.GetLikes()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error while fetching likes",
				"error":   err.Error(),
			})
		}
		return c.JSON(http.StatusOK, likes)
	}
}

func getLike(service domain.LikeService) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		likes, err := service.GetLike(id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "likes not found",
				"error":   err.Error(),
			})
		}
		return c.JSON(http.StatusOK, likes)
	}
}

func createLike(service domain.LikeService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var likeData struct {
			CommentId string `json:"commentId"`
		}
		if err := c.Bind(&likeData); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid request payload",
				"error":   err.Error(),
			})
		}
		like, err := service.CreateLike(likeData.CommentId)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error while creating like",
				"error":   err.Error(),
			})
		}
		return c.JSON(http.StatusCreated, like)
	}
}

func main() {
	e := echo.New()

	likeRepo := &domain.InMemoryLikeRepository{
		Like: make(map[string]*domain.Like),
	}
	likeService := &domain.LikeServiceImpl{LikeRepository: likeRepo}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	api := e.Group("/api")
	api.GET("/likes", getLikes(likeService))
	api.GET("/likes/:id", getLike(likeService))
	api.POST("/likes", createLike(likeService))

	// WebSocket endpoint
	e.GET("/ws/likes", likesHandler(likeService))

	e.Logger.Fatal(e.Start(":8080"))
}
