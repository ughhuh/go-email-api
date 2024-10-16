package main

import (
	"github.com/gin-gonic/gin"
)

func main() {
	// setup
	router := gin.Default()
	// get emails for email inbox
	router.GET("/inbox/:email_id", getEmailsForUser)
	router.GET("/email/:email_id", getEmailById)

	router.Run(":8080")
}
