package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/szluyu99/rabbit"
)

var fileMode os.FileMode = 0666

var (
	serverAddr string
	logFile    string
	dbDriver   string
	dsn        string
)

func main() {
	flag.StringVar(&serverAddr, "s", ":8080", "listen addr")
	flag.StringVar(&logFile, "l", "", "log file")
	flag.StringVar(&dbDriver, "d", "", "DB Driver, sqlite|mysql")
	flag.StringVar(&dsn, "n", "", "DB DSN")
	flag.Parse()

	var err error
	var lw io.Writer
	if logFile != "" {
		lw, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, fileMode)
		if err != nil {
			log.Fatalf("open %s fail, %v\n", logFile, err)
		}
	}

	// logger with color
	rabbit.EnabledConsoleColor = true
	rabbit.Infoln("Server started", "addr", serverAddr)

	// TEST_KEY=100 go run main.go
	key := rabbit.GetEnv("TEST_KEY")
	if key != "" {
		rabbit.Infoln("TEST_KEY", key)
	} else {
		rabbit.Errorln("TEST_KEY not found")
	}

	// db
	db := rabbit.InitDatabase(dbDriver, dsn, lw)

	// router
	r := gin.New()
	r.Use(gin.LoggerWithWriter(lw), gin.Recovery())

	// init rabbit
	rabbit.InitRabbit(db, r)

	// register handlers
	ar := r.Group("/api").Use(rabbit.WithAuthentication(), rabbit.WithAuthorization("/api"))
	rabbit.RegisterAuthorizationHandlers(db, ar)

	r.Run(serverAddr)
}
