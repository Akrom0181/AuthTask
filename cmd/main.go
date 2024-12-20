package main

import (
	"net/http"
	"os"

	// "task/api"
	"task/api"
	"task/config"
	"task/pkg/logger"
	"task/service"

	postgres "task/storage/postgres"
	"task/storage/redis"

	"github.com/gin-gonic/gin"
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

	*loggerLevel = logger.LevelInfo

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

	api.NewApi(r, &cfg, pgconn, log, services)

	// go KeepAlive(&cfg)

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// fmt.Println("Listening server", os.Getenv("POSTGRES_HOST")+os.Getenv("HTTP_PORT"))
	err = r.Run(os.Getenv("HTTP_PORT"))
	if err != nil {
		panic(err)
	}

}
