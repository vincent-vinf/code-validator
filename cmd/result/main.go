package main

import (
	"flag"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/vincent-vinf/go-jsend"

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
	r.Use(util.Cors())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Page not found"})
	})
	r.GET("/metrics", util.PrometheusHandler())

	authMiddleware, err := jwtx.GetAuthMiddleware(cfg.JWT.Secret, cfg.JWT.Timeout, cfg.JWT.MaxRefresh)
	if err != nil {
		log.Fatal(err)
	}

	router := r.Group(util.WithGlobalAPIPrefix("/result"))
	router.Use(authMiddleware.MiddlewareFunc())
	router.GET("/:id", getTaskDetailByID)
	router.GET("", getResultList)

	util.WatchSignalGrace(r, *port)
}

func getTaskDetailByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	task, err := db.GetTaskInfoByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, jsend.SimpleErr(err.Error()))
		return
	}
	c.JSON(http.StatusOK, jsend.Success(task))
}

func getResultList(c *gin.Context) {
	t, _ := c.Get(jwtx.IdentityKey)
	user := t.(*jwtx.TokenUserInfo)
	log.Info(user.ID)

	batchID, _ := strconv.Atoi(c.Query("batch"))
	tasks, err := db.ListTasks(batchID, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, jsend.SimpleErr(err.Error()))
		return
	}
	c.JSON(http.StatusOK, jsend.Success(tasks))
}
