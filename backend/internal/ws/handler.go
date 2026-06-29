package ws

import (
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"optistock/internal/auth"
)

func Handler(tokenManager *auth.TokenManager, hub *Hub) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		token := strings.TrimSpace(conn.Query("token"))
		if token == "" {
			authHeader := conn.Headers("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			}
		}
		claims, err := tokenManager.ParseAccessToken(token)
		if err != nil {
			_ = conn.Close()
			return
		}

		client := &Client{
			UserID: claims.UserID,
			Role:   string(claims.Role),
			Conn:   conn,
		}
		hub.Register(client)
		defer func() {
			hub.Unregister(client)
			_ = conn.Close()
		}()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	})
}
