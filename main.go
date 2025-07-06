package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"scripts-llm/internal/chat"
	"scripts-llm/internal/role"
	"scripts-llm/internal/script"
)

func main() {
	s := gin.Default()
	g := s.Group("/api/v1")

	g.POST("/scripts", script.Upload)
	g.POST("/roles", script.Analyze)
	g.GET("/:script_id/roles", role.Query)
	g.PUT("/roles", role.Update)
	g.POST("/chat", chat.SSE)
	g.GET("/chat", chat.History)

	if err := s.Run("0.0.0.0:8080"); err != nil {
		log.Println(err, "main")
	}
}
