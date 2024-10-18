package main

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// setup
	configLoader("config.json") // load config
	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// get emails for email inbox
	router.GET("/inbox/:email_id", getEmailsForUser)
	router.GET("/email/:email_id", getEmailById)

	router.POST("/email", createNewInbox)
	router.DELETE("/email", deleteInbox)

	port := ":" + viper.GetString("PORT")
	router.Run(port)
}

func configLoader(configName string) {
	viper.SetConfigName(configName)
	viper.SetConfigType(strings.Split(configName, ".")[1])
	viper.AddConfigPath(".")     // look in current dir
	viper.AddConfigPath("../..") // look in parent dir
	err := viper.ReadInConfig()  // red the config file
	if err != nil {
		panic(fmt.Errorf("error when reading config file: %s", err))
	}
	// watch config for changes
	viper.WatchConfig()
}
