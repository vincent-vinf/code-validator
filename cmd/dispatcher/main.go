package main

import (
	"flag"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/vincent-vinf/code-validator/pkg/orm"
	"github.com/vincent-vinf/code-validator/pkg/util"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"github.com/vincent-vinf/code-validator/pkg/util/db"
	"github.com/vincent-vinf/code-validator/pkg/util/mq"
)

var (
	configPath = flag.String("config-path", "configs/config.yaml", "")
	log        = logrus.New()
	port       = flag.Int("port", 8001, "")

	pubClient *mq.PubClient
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
	//ossClient, err := oss.NewClient(cfg)
	//if err != nil {
	//	log.Fatal(err)
	//}

	pubClient, err = mq.NewPubClient(cfg.RabbitMQ)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.New()
	//gin.SetMode(gin.ReleaseMode)
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Page not found"})
	})

	//authMiddleware, err := jwtx.GetAuthMiddleware(cfg.JWT.Secret, cfg.JWT.Timeout, cfg.JWT.MaxRefresh)
	//if err != nil {
	//	log.Fatal(err)
	//}

	router := r.Group("/batch")
	//router.Use(authMiddleware.MiddlewareFunc())
	router.GET("/:id", getBatchByID)
	router.GET("", getBatchList)
	router.POST("", addBatch)
	router.GET("/token", addBatchToken)
	router.POST("/token/:uid/upload", upload)

	router.GET("/task/:id", getTaskByID)
	router.GET("/:id/task", getTaskByBatchID)
	router.POST("/:id/task", addTaskOfBatch)

	util.WatchSignalGrace(r, *port)
}

func getBatchByID(c *gin.Context) {
	id := c.Param("id")
	log.Info(id)
}

func getBatchList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "2"})
}

func addBatch(c *gin.Context) {
	batch := &orm.Batch{}
	if err := c.BindJSON(batch); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})

		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "1"})
}

func addBatchToken(c *gin.Context) {
	u := uuid.New()
	c.JSON(http.StatusOK, gin.H{"message": u.String()})
}

func upload(c *gin.Context) {
	var err error
	defer func() {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}()
	uid := c.Param("uid")

	c.Query("key")

	path := c.Param("path")
	file, err := c.FormFile("file")
	if err != nil {
		return
	}
	log.Println(file.Filename)
	path = filepath.Clean(path)

	//todo fix filename
	err = c.SaveUploadedFile(file, filepath.Join(uid, path))
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": file.Filename})
}

func getTaskByID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "2"})
}

func getTaskByBatchID(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "2"})
}

func addTaskOfBatch(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "3"})
}
