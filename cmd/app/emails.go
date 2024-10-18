package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/spf13/viper"
)

type ArrayResponse struct {
	Data []Email `json:"data"`
}

type User struct {
	Address string `json:"email_address"`
}

type Email struct {
	MessageId string   `json:"message_id"`
	Body      string   `json:"body"`
	From      []string `json:"from"`
	To        []string `json:"to"`
}

type CreateEmailRequest struct {
	Username string `json:"username"`
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

func createNewInbox(c *gin.Context) {
	// perhaps i should allow to custom make
	var requestBody CreateEmailRequest
	json.NewDecoder(c.Request.Body).Decode(&requestBody)
	// so we generate a new email inbox
	if requestBody.Username == "" {
		requestBody.Username = uuid.NewString()
	}
	allowedDomains := viper.GetStringSlice("allowed_domains")
	domainIndex := rand.Intn(len(allowedDomains))
	username := requestBody.Username + "@" + allowedDomains[domainIndex]
	// post new inbox
	db := connectToDb()
	query := `INSERT INTO users("email_address") values ($1);`
	_, err := db.Exec(query, username)
	db.Close()
	if err != nil {
		log.Fatal(err)
		c.JSON(500, "Failed to create email address")
		return
	}
	// send new inbox to user
	c.JSON(201, User{Address: username})
}

func deleteInbox(c *gin.Context) {
	// get email address from body
	var requestBody User
	json.NewDecoder(c.Request.Body).Decode(&requestBody)
	if requestBody.Address == "" {
		c.JSON(400, "Invalid email")
		return
	}
	// get emails
	db := connectToDb()
	query := `select mail_id from inboxes where user_id = $1`
	rows, err := db.Query(query, requestBody.Address)
	db.Close()
	if err != nil {
		log.Fatal(err)
	}
	emails := []string{}
	for rows.Next() {
		var entry string
		err = rows.Scan(&entry)
		if err != nil {
			log.Fatal(err)
			return
		}
		emails = append(emails, entry)
	}
	// delete user and inboxes
	db = connectToDb()
	query = `DELETE FROM users WHERE email_address = $1;`
	_, err = db.Exec(query, requestBody.Address)
	db.Close()
	// return on success
	if err != nil {
		log.Fatal(err)
		c.JSON(500, "Failed to delete email address")
		return
	}
	// delete email entries
	db = connectToDb()
	emails_str := fmt.Sprint(strings.Join(emails, ","))
	query = `delete from emails where message_id in ($1)`
	_, err = db.Exec(query, emails_str)
	db.Close()
	if err != nil {
		log.Fatal(err)
		return
	}
	c.JSON(200, "all good")
}
