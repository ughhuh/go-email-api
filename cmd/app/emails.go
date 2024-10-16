package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type ArrayResponse struct {
	Data []Email `json:"data"`
}

type Email struct {
	MessageId string
	Body      string
	From      []string
	To        []string
}

func getEmailsForUser(c *gin.Context) {
	// get id from uri
	email := c.Param("email_id")
	db := connectToDb()
	query := `select message_id, body from emails where message_id in (select mail_id from inboxes where user_id = $1);`
	rows, err := db.Query(query, email)
	db.Close()
	if err != nil {
		log.Fatal(err)
	}
	emails := []Email{}
	for rows.Next() {
		var entry Email
		err = rows.Scan(&entry.MessageId, &entry.Body)
		if err != nil {
			log.Fatal(err)
		}
		emails = append(emails, entry)
	}

	c.JSON(200, ArrayResponse{Data: emails})
}

func getEmailById(c *gin.Context) {
	emailID := c.Param("email_id")
	db := connectToDb()
	query := `SELECT message_id, body, "from", "to" FROM emails WHERE message_id = $1`
	rows, err := db.Query(query, emailID)
	db.Close()
	if err != nil {
		log.Fatal(err)
	}
	var email Email
	for rows.Next() {
		err = rows.Scan(&email.MessageId, &email.Body, pq.Array(&email.From), pq.Array(&email.To))
		if err != nil {
			log.Fatal(err)
		}
	}
	c.JSON(200, email)
}
