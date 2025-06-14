package main

import "github.com/gin-gonic/gin"

func main() {
	router := gin.Default()
	registerRoutes(router)
	router.Run(":9001")
}
