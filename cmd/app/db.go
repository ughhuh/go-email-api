package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)

var queries_map = map[string]string{
	"getSimpleEmailByMsgId": `SELECT message_id, "from", date FROM emails WHERE message_id IN (SELECT mail_id FROM inboxes WHERE user_id = $1);`,
	"getEmailByMsgId":       `SELECT message_id, body, "from", "to" FROM emails WHERE message_id = $1`,
	"getMsgIdByUserId":      `select mail_id from inboxes where user_id = $1`,
	"createNewUser":         `INSERT INTO users(email_address) values ($1);`,
	"deleteUserByUserId":    `DELETE FROM users WHERE email_address = $1;`,
	"deleteMsgByMsgId":      `delete from emails where message_id = $1`,
}

func connectToDb() *sql.DB {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("DB_HOST"), viper.GetString("DB_USER"), viper.GetString("DB_SECRET"), viper.GetString("DB_NAME"), viper.GetString("ssl_mode"))
	fmt.Print(connStr)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

// so on init we can call a function that creates an object that prepares queries
// then those queries are fetched by router functions and executed there
