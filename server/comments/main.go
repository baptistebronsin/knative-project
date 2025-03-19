package main

import (
	"bytes"
	"comments/domain"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func commentsHandler(service domain.CommentService) echo.HandlerFunc {
	return func(c echo.Context) error {
		conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			println("Upgrade error:", err)
			return err
		}
		defer conn.Close()

		println("WebSocket connection established on /comments")

		sendComments := func() {
			comments, err := service.GetComments()
			if err != nil {
				println("Error fetching comments:", err)
				conn.WriteMessage(websocket.TextMessage, []byte(`{"error": "Failed to retrieve comments"}`))
				return
			}

			data, _ := json.Marshal(comments)
			conn.WriteMessage(websocket.TextMessage, data)
		}

		sendComments()
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sendComments()
			}
		}
	}
}

func getComments(service domain.CommentService) echo.HandlerFunc {
	return func(c echo.Context) error {
		comments, err := service.GetComments()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error while fetching comments",
				"error":   err.Error(),
			})
		}
		return c.JSON(http.StatusOK, comments)
	}
}

func getComment(service domain.CommentService) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		comment, err := service.GetComment(id)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"message": "Comment not found",
				"error":   err.Error(),
			})
		}
		return c.JSON(http.StatusOK, comment)
	}
}

func createComment(service domain.CommentService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var commentData struct {
			Content string `json:"content"`
			Emotion string `json:"emotion"`
		}
		if err := c.Bind(&commentData); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid request payload",
				"error":   err.Error(),
			})
		}
		comment, err := service.CreateComment(commentData.Content, commentData.Emotion)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error while creating comment",
				"error":   err.Error(),
			})
		}
		return c.JSON(http.StatusCreated, comment)
	}
}

func createCommentBroker(service domain.CommentService) echo.HandlerFunc {
	return func(c echo.Context) error {
		var event struct {
			Type string          `json:"type"`
			Data json.RawMessage `json:"data"`
		}

		if err := c.Bind(&event); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Invalid request payload",
				"error":   err.Error(),
			})
		}

		log.Printf("Received event type: %s", event.Type)

		if event.Type != "new-review-comment" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"message": "Unexpected event type",
			})
		}

		brokerURI := os.Getenv("K_SINK")
		if brokerURI == "" {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Broker URI not configured",
			})
		}

		eventData := map[string]interface{}{
			"specversion": "1.0",
			"type":        "new-review-comment",
			"source":      "bookstore-eda",
			"id":          fmt.Sprintf("%d", time.Now().UnixNano()),
			"data":        event.Data,
		}

		jsonData, err := json.Marshal(eventData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error marshalling event data",
				"error":   err.Error(),
			})
		}

		req, err := http.NewRequest("POST", brokerURI, bytes.NewBuffer(jsonData))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Error creating request",
				"error":   err.Error(),
			})
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("ce-specversion", "1.0")
		req.Header.Set("ce-type", "new-review-comment")
		req.Header.Set("ce-source", "bookstore-eda")
		req.Header.Set("ce-id", fmt.Sprintf("%d", time.Now().UnixNano()))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "Failed to forward event",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"success": "true",
			"message": "Event forwarded successfully",
		})
	}
}

func main() {
	e := echo.New()

	commentRepo := &domain.InMemoryCommentRepository{
		Comment: make(map[string]*domain.Comment),
	}
	commentService := &domain.CommentServiceImpl{CommentRepository: commentRepo}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	api := e.Group("/api")
	api.GET("/comments", getComments(commentService))
	api.GET("/comments/:id", getComment(commentService))
	api.POST("/comments", createComment(commentService))
	api.POST("/comments/broker", createCommentBroker(commentService))

	// WebSocket endpoint
	e.GET("/ws/comments", commentsHandler(commentService))

	e.Logger.Fatal(e.Start(":8080"))
}
