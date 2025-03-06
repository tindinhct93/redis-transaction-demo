package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/yourusername/redis-demo/backend/redistest"
)

var ctx = context.Background()

func main() {
	r := gin.Default()

	r.GET("/txpipeline", redistest.Main)
	r.Run(":8080")
}
