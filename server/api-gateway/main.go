package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type HttpVerb string

const (
	GET    HttpVerb = "GET"
	POST   HttpVerb = "POST"
	PUT    HttpVerb = "PUT"
	PATCH  HttpVerb = "PATCH"
	DELETE HttpVerb = "DELETE"
)

var upgrader_comment = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (for development purposes)
		return true
	},
}

var upgrader_like = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from any origin (for development purposes)
		return true
	},
}

func callService(c echo.Context, httpVerb HttpVerb, url string, body io.Reader) error {
	req, err := http.NewRequest(string(httpVerb), url, body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Error creating request",
			"error":   err.Error(),
		})
	}

	if httpVerb == POST || httpVerb == PUT || httpVerb == PATCH {
		req.Header.Set("Content-Type", "application/json")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Error sending request",
			"error":   err.Error(),
		})
	}
	defer resp.Body.Close()

	// Parse response body
	bodyResponse, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Error reading response body",
			"error":   err.Error(),
		})
	}

	var bodyFormatted interface{}
	err = json.Unmarshal(bodyResponse, &bodyFormatted)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "Error parsing response body",
			"error":   err.Error(),
		})
	}

	return c.JSON(resp.StatusCode, bodyFormatted)
}

func proxyWebSocket(target string, upgrader websocket.Upgrader) echo.HandlerFunc {
	url, err := url.Parse(target)
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	return func(c echo.Context) error {
		r := c.Request()
		w := c.Response()

		if strings.Contains(r.Header.Get("Connection"), "Upgrade") && r.Header.Get("Upgrade") == "websocket" {
			// Upgrade the connection to WebSocket
			conn, err := upgrader.Upgrade(w, r, nil)
			if err != nil {
				log.Println("Upgrade error:", err)
				return err
			}
			defer conn.Close()

			// Proxy the WebSocket connection
			backendConn, _, err := websocket.DefaultDialer.Dial(target, nil)
			if err != nil {
				log.Println("Dial error:", err)
				return err
			}
			defer backendConn.Close()

			// Copy messages between client and backend
			done := make(chan struct{})
			go func() {
				defer close(done)
				for {
					_, msg, err := conn.ReadMessage()
					if err != nil {
						log.Println("Read error:", err)
						return
					}
					if err := backendConn.WriteMessage(websocket.TextMessage, msg); err != nil {
						log.Println("Write error:", err)
						return
					}
				}
			}()

			go func() {
				for {
					_, msg, err := backendConn.ReadMessage()
					if err != nil {
						log.Println("Read error:", err)
						return
					}
					if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
						log.Println("Write error:", err)
						return
					}
				}
			}()

			<-done
			return nil
		} else {
			// Handle non-WebSocket requests
			proxy.ServeHTTP(w, r)
			return nil
		}
	}
}

const COMMENTS_API_URL = "http://bookstore-api-comments-svc:8080/api/comments"
const COMMENTS_WS_URL = "ws://bookstore-api-comments-svc:8080/ws/comments"

const LIKES_API_URL = "http://bookstore-api-likes-svc:8080/api/likes"
const LIKES_WS_URL = "ws://bookstore-api-likes-svc:8080/ws/likes"

func getComments(c echo.Context) error {
	return callService(c, GET, COMMENTS_API_URL, nil)
}

func getComment(c echo.Context) error {
	return callService(c, GET, COMMENTS_API_URL+"/"+c.Param("id"), nil)
}

func createComment(c echo.Context) error {
	return callService(c, POST, COMMENTS_API_URL, c.Request().Body)
}

func createCommentBroker(c echo.Context) error {
	return callService(c, POST, COMMENTS_API_URL+"/broker", c.Request().Body)
}

func getLikes(c echo.Context) error {
	return callService(c, GET, LIKES_API_URL, nil)
}

func getLike(c echo.Context) error {
	return callService(c, GET, LIKES_API_URL+"/"+c.Param("id"), nil)
}

func createLike(c echo.Context) error {
	return callService(c, POST, LIKES_API_URL, c.Request().Body)
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))

	// Rest API endpoints
	api := e.Group("/api")
	api.GET("/comments", getComments)
	api.GET("/comments/:id", getComment)
	api.POST("/comments", createComment)
	api.POST("/comments/broker", createCommentBroker)

	api.GET("/likes", getLikes)
	api.GET("/likes/:id", getLike)
	api.POST("/likes", createLike)

	// Websocket endpoints
	e.GET("/ws/comments", proxyWebSocket(COMMENTS_WS_URL, upgrader_comment))
	e.GET("/ws/likes", proxyWebSocket(LIKES_WS_URL, upgrader_like))

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}
