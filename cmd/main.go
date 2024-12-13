package main

import (
	"fmt"
	"os"

	// "task/api"
	"task/api"
	"task/config"
	"task/pkg/logger"
	"task/service"

	postgres "task/storage/postgres"
	"task/storage/redis"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// func KeepAlive(cfg *config.Config) {
// 	for {
// 		_, err := http.Get(fmt.Sprintf("http://localhost%s/ping", cfg.HTTPPort))
// 		if err != nil {
// 			fmt.Println("Error while sending ping:", err)
// 		} else {
// 			fmt.Println("Ping sent successfully")
// 		}
// 		time.Sleep(1 * time.Minute)
// 	}
// }

func main() {
	cfg := config.Load()

	var loggerLevel = new(string)

	*loggerLevel = logger.LevelDebug

	switch cfg.Environment {
	case config.DebugMode:
		*loggerLevel = logger.LevelDebug
		gin.SetMode(gin.DebugMode)
	case config.TestMode:
		*loggerLevel = logger.LevelDebug
		gin.SetMode(gin.TestMode)
	default:
		*loggerLevel = logger.LevelInfo
		gin.SetMode(gin.ReleaseMode)
	}

	log := logger.NewLogger("app", *loggerLevel)
	defer func() {
		err := logger.Cleanup(log)
		if err != nil {
			return
		}
	}()

	pgconn, err := postgres.NewConnectionPostgres(&cfg)
	if err != nil {
		panic("postgres no connection: " + err.Error())
	}
	defer pgconn.CloseDB()

	newRedis := redis.New(cfg)
	services := service.New(pgconn, log, newRedis)

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())

	// Set trusted proxies
	err = r.SetTrustedProxies([]string{
		"13.228.225.19",
		"18.142.128.26",
		"54.254.162.138",
	})
	if err != nil {
		log.Fatal("Failed to set trusted proxies: ", zap.Error(err))
	}

	api.NewApi(r, &cfg, pgconn, log, services)

	// Log the server start
	fmt.Println("Listening server on", os.Getenv("POSTGRES_HOST")+os.Getenv("HTTP_PORT"))
	err = r.Run(os.Getenv("HTTP_PORT"))
	if err != nil {
		panic(err)
	}
}

// go KeepAlive(&cfg)

// r.GET("/ping", func(c *gin.Context) {
// 	c.String(200, "pong")
// })
