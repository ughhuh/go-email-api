package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	// Define flags
	pflag.String("config", "config.json", "Path to configuration file")
	pflag.String("logdir", "./logs", "Directory for log files")

	// Bind command-line flags to pflag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()

	// Bind flags to viper
	viper.BindPFlags(pflag.CommandLine)

	ensureLogDirectory(viper.GetString("logdir"))
	cfile := strings.Split(viper.GetString("config"), ".")

	configLoader(cfile[0], cfile[1])

	// set mode
	gin.SetMode(viper.GetString("mode"))

	// Initialize database
	db := connectToDb()
	defer db.Close()

	router := gin.New()
	router.SetTrustedProxies(viper.GetStringSlice("trusted_proxies"))

	if viper.IsSet("logrotate") {
		logFile := &lumberjack.Logger{
			Filename:   viper.GetString("logrotate.log_file"),
			MaxSize:    viper.GetInt("logrotate.max_size"),
			MaxBackups: viper.GetInt("logrotate.max_backups"),
			MaxAge:     viper.GetInt("logrotate.max_age"),
			Compress:   viper.GetBool("logrotate.compress"),
		}

		ensureLogFile(viper.GetString("logrotate.log_file"))

		// set output to both console and log rotator
		multiWriter := io.MultiWriter(logFile, os.Stdout)

		// Set the log output to the multi-writer
		log.SetOutput(multiWriter)

		router.Use(gin.LoggerWithWriter(multiWriter))
		router.Use(gin.RecoveryWithWriter(multiWriter))
	} else {
		router.Use(gin.LoggerWithWriter(log.Writer()))
		router.Use(gin.RecoveryWithWriter(log.Writer()))
	}

	// set middleware
	router.Use(SetDatabase(db))
	router.Use(PrepareSQLQueries())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// db ping every 10 seconds
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Ctx cancelled")
				return
			default:
				if err := db.Ping(); err != nil {
					log.Printf("Database connection is down: %v", err)
				}
				time.Sleep(10 * time.Second)
			}
		}
	}(ctx)

	// set routers
	router.GET("/inbox/:address", getEmailsForUser)
	router.GET("/email/:message_id", getEmailById)
	router.POST("/email", createNewInbox)
	router.DELETE("/email", deleteInbox)

	// start listening on config port
	port := ":" + viper.GetString("PORT")
	srv := &http.Server{
		Addr:    port,
		Handler: router,
	}

	// following https://github.com/gin-gonic/examples/blob/master/graceful-shutdown/graceful-shutdown/notify-with-context/server.go
	// init server as goroutine and listen for exceptions on main thread
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %s\n", err)
		}
	}()

	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel,
		syscall.SIGHUP,
		syscall.SIGQUIT,
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGABRT)

	for signal := range signalChannel {
		switch signal {
		case syscall.SIGHUP:
			// reload jk
			log.Println("Caught hangup")
		default:
			// The context is used to inform the server it has 5 seconds to finish
			// the request it is currently processing
			log.Println("Shutting down the server.")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Fatal("Server forced to shutdown: ", err)
			}
			return
		}
	}
}

func configLoader(configName string, configType string) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")    // look in current dir
	err := viper.ReadInConfig() // read env file
	if err != nil {
		log.Printf("Warning: Could not read .env file: %s", err)
	}
	viper.AutomaticEnv()

	viper.SetConfigName(configName)
	viper.SetConfigType(configType)

	err = viper.MergeInConfig() // read the config file
	if err != nil {
		log.Panicf("Error when reading config file: %s", err)
	}
	// watch config for changes
	viper.WatchConfig()
}

// middleware

// global database setter
func SetDatabase(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

// sql query preparing
func PrepareSQLQueries() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(*sql.DB)
		for key, query := range queries_map {
			cache, err := db.Prepare(query)
			if err != nil {
				log.Panicf("Error when preparing a query: %s", err)
			}
			c.Set(key, cache)
		}
	}
}

func ensureLogDirectory(logdir string) {
	if _, err := os.Stat(logdir); os.IsNotExist(err) {
		err := os.Mkdir(logdir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create log directory: %v", err)
		}
	}
}

func ensureLogFile(filename string) {
	// Check if the file exists
	if _, err := os.Stat(filename); err != nil {
		if os.IsNotExist(err) {
			// File does not exist, create it
			file, err := os.Create(filename)
			if err != nil {
				log.Fatal("failed to create file: %w", err)
			}
			defer file.Close()
		} else {
			log.Fatal("error checking file: %w", err)
		}
	}
}
