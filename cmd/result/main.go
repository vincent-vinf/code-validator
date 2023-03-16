package main

import (
	"flag"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/vincent-vinf/code-validator/pkg/util"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"github.com/vincent-vinf/code-validator/pkg/util/db"
	"github.com/vincent-vinf/code-validator/pkg/util/jwtx"
)

var (
	configPath = flag.String("config-path", "configs/config.yaml", "")
	log        = logrus.New()
	port       = flag.Int("port", 8003, "")
)

func init() {
	flag.Parse()
}

func main() {
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	db.Init(cfg.Mysql)
	defer db.Close()

	r := gin.New()
	//gin.SetMode(gin.ReleaseMode)
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Page not found"})
	})

	authMiddleware, err := jwtx.GetAuthMiddleware(cfg.JWT.Secret, cfg.JWT.Timeout, cfg.JWT.MaxRefresh)
	if err != nil {
		log.Fatal(err)
	}

	router := r.Group("/result")
	router.Use(authMiddleware.MiddlewareFunc())

	util.WatchSignalGrace(r, *port)
}
