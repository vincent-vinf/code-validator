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
	port       = flag.Int("port", 8002, "")
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
	r.Use(util.Cors())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Page not found"})
	})

	authMiddleware, err := jwtx.GetAuthMiddleware(cfg.JWT.Secret, cfg.JWT.Timeout, cfg.JWT.MaxRefresh)
	if err != nil {
		log.Fatal(err)
	}
	r.GET("/metrics", util.PrometheusHandler())

	router := r.Group(util.WithGlobalAPIPrefix("/user"))
	router.POST("/login", authMiddleware.LoginHandler)
	router.POST("/register", registerHandler)
	// 一组需要验证的路由
	auth := router.Group("/auth")
	auth.Use(authMiddleware.MiddlewareFunc())
	auth.GET("/refresh_token", authMiddleware.RefreshHandler)

	util.WatchSignalGrace(r, *port)
}

func registerHandler(c *gin.Context) {
	var registerForm jwtx.RegisterForm
	if err := c.ShouldBind(&registerForm); err != nil {
		c.JSON(400, gin.H{
			"error": "Bad request parameter",
		})
		return
	}

	isExist, err := db.IsExistEmail(registerForm.Email)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"error": "Server internal error",
		})
		return
	}
	if isExist {
		c.JSON(400, gin.H{
			"error": "email already exists",
		})
		return
	}

	id, err := db.Register(registerForm.Username, registerForm.Email, registerForm.Passwd)
	if err != nil {
		log.Println(err)
		c.JSON(500, gin.H{
			"error": "Server internal error",
		})
		return
	}

	c.JSON(200, gin.H{
		"token": jwtx.GenerateToken(id),
	})
}
