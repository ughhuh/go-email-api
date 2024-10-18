package main

import (
	"encoding/json"
	"log"
	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/spf13/viper"
)

type User struct {
	Address string `json:"email_address"`
}

type Email struct {
	MessageId string   `json:"message_id"`
	Body      string   `json:"body"`
	From      []string `json:"from"`
	To        []string `json:"to"`
}

type SimpleEmail struct {
	MessageId string `json:"message_id"`
	From      string `json:"from"`
	Date      string `json:"date"`
}

type CreateEmailRequest struct {
	Username string `json:"username"`
}

func getEmailsForUser(c *gin.Context) {
	// get id from uri
	email := c.Param("email_id")
	db := connectToDb()
	defer db.Close()
	query := `select message_id, "from", date" from emails where message_id in (select mail_id from inboxes where user_id = $1);`
	rows, err := db.Query(query, email)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve emails."})
		return
	}
	defer rows.Close()
	emails := []SimpleEmail{}
	for rows.Next() {
		var entry SimpleEmail
		err = rows.Scan(&entry.MessageId, &entry.From, &entry.Date)
		if err != nil {
			c.JSON(500, gin.H{"error": "An error occured while parsing email records."})
			return
		}
		emails = append(emails, entry)
	}

	c.JSON(200, gin.H{"emails": emails})
}

func getEmailById(c *gin.Context) {
	emailID := c.Param("email_id")
	db := connectToDb()
	defer db.Close()
	query := `SELECT message_id, body, "from", "to" FROM emails WHERE message_id = $1`
	rows, err := db.Query(query, emailID)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve emails."})
		return
	}
	defer rows.Close()
	var email Email
	for rows.Next() {
		err = rows.Scan(&email.MessageId, &email.Body, pq.Array(&email.From), pq.Array(&email.To))
		if err != nil {
			c.JSON(500, gin.H{"error": "An error occured while parsing email records."})
			return
		}
	}
	c.JSON(200, gin.H{"email": email})
}

func createNewInbox(c *gin.Context) {
	// perhaps i should allow to custom make
	var requestBody CreateEmailRequest
	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}
	if requestBody.Username == "" {
		requestBody.Username = uuid.NewString()
	}
	allowedDomains := viper.GetStringSlice("allowed_domains")
	domainIndex := rand.Intn(len(allowedDomains))
	username := requestBody.Username + "@" + allowedDomains[domainIndex]
	// post new inbox
	db := connectToDb()
	defer db.Close()
	query := `INSERT INTO users(email_address) values ($1);`
	_, err = db.Exec(query, username)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create new email address."})
		return
	}
	// send new inbox to user
	c.JSON(201, gin.H{"email": username})
}

func deleteInbox(c *gin.Context) {
	// get email address from body
	var requestBody User
	err := json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format"})
		return
	}
	if requestBody.Address == "" {
		c.JSON(400, gin.H{"error": "Invalid email."})
		return
	}
	// get emails
	db := connectToDb()
	defer db.Close()

	query := `select mail_id from inboxes where user_id = $1`
	rows, err := db.Query(query, requestBody.Address)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	emails := []string{}
	for rows.Next() {
		var entry string
		err = rows.Scan(&entry)
		if err != nil {
			c.JSON(500, gin.H{"error": "An error occured while parsing email records."})
			return
		}
		emails = append(emails, entry)
	}
	// delete user and inboxes
	query = `DELETE FROM users WHERE email_address = $1;`
	_, err = db.Exec(query, requestBody.Address)

	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to delete email user."})
		return
	}
	// delete email entries
	if len(emails) > 0 {
		query = `delete from emails where message_id in ANY($1)`
		_, err = db.Exec(query, pq.Array(emails))
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to delete user's emails."})
			return
		}
	}

	c.JSON(200, gin.H{"message": "User and associated emails deleted successfully"})
}
