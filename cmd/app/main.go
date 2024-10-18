package main

import (
	"database/sql"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func main() {
	// setup
	configLoader("config", "json") // load config

	router := gin.Default()
	router.SetTrustedProxies(viper.GetStringSlice("trusted_proxies"))

	// set middleware
	router.Use(SetDatabase())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// set routers
	router.GET("/inbox/:email_id", getEmailsForUser)
	router.GET("/email/:email_id", getEmailById)
	router.POST("/email", createNewInbox)
	router.DELETE("/email", deleteInbox)

	// start listening on config port
	port := ":" + viper.GetString("PORT")
	router.Run(port)
}

func configLoader(configName string, configType string) {
	viper.SetConfigName(".env")
	viper.AddConfigPath(".")     // look in current dir
	viper.AddConfigPath("../..") // look in parent dir
	viper.ReadInConfig()         // read env file

	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	err := viper.ReadInConfig() // read the config file
	if err != nil {
		panic(fmt.Errorf("error when reading config file: %s", err))
	}
	// watch config for changes
	viper.WatchConfig()
}

// global database setter
func SetDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := connectToDb()
		defer db.Close()
		c.Set("db", db)
	}
}

// sql query preparing
func PrepareSQLQueries() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(*sql.DB)
		for key, query := range queries_map {
			cache, err := db.Prepare(query)
			if err != nil {
				panic(fmt.Errorf("error when preparing a query: %s", err))
			}
			c.Set(key, cache)
		}
	}
}
