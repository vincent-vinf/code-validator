package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vincent-vinf/code-validator/pkg/util/jwtx"
	"github.com/vincent-vinf/code-validator/pkg/util/oss"
	"net/http"
	"path"
	"strconv"

	"github.com/vincent-vinf/code-validator/pkg/orm"
	"github.com/vincent-vinf/code-validator/pkg/util"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"github.com/vincent-vinf/code-validator/pkg/util/db"
	"github.com/vincent-vinf/code-validator/pkg/util/mq"
	"github.com/vincent-vinf/go-jsend"
)

const (
	defaultTmpDir      = "tmp"
	defaultContentType = gin.MIMEPlain
)

var (
	configPath = flag.String("config-path", "configs/config.yaml", "")
	log        = logrus.New()
	port       = flag.Int("port", 8001, "")

	pubClient *mq.PubClient
	ossClient *oss.Client
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

	ossClient, err = oss.NewClient(cfg.Minio)
	if err != nil {
		log.Fatal(err)
	}

	pubClient, err = mq.NewPubClient(cfg.RabbitMQ)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.New()
	//gin.SetMode(gin.ReleaseMode)
	r.Use(gin.Logger())
	r.Use(util.Cors())
	r.Use(gin.Recovery())

	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Page not found"})
	})

	authMiddleware, err := jwtx.GetAuthMiddleware(cfg.JWT.Secret, cfg.JWT.Timeout, cfg.JWT.MaxRefresh)
	if err != nil {
		log.Fatal(err)
	}

	router := r.Group("/batch")
	router.Use(authMiddleware.MiddlewareFunc())
	router.GET("/:id", getBatchByID)
	router.GET("", getBatchList)
	router.POST("", addBatch)
	router.POST("/file", uploadFile)

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
	c.JSON(http.StatusOK, gin.H{"data": "2"})
}

func addBatch(c *gin.Context) {
	batch := &orm.Batch{}
	if err := c.BindJSON(batch); err != nil {
		c.JSON(400, gin.H{"message": err.Error()})

		return
	}
	c.JSON(http.StatusOK, gin.H{"data": "1"})
}

func uploadFile(c *gin.Context) {
	t, _ := c.Get(jwtx.IdentityKey)
	user := t.(*jwtx.TokenUserInfo)

	var err error
	defer func() {
		if err != nil {
			c.JSON(http.StatusInternalServerError, jsend.SimpleErr(err.Error()))
		}
	}()

	file, err := c.FormFile("file")
	if err != nil {
		return
	}
	log.Info("filename:", file.Filename)

	contentType := defaultContentType
	if len(file.Header["Content-Type"]) > 0 {
		contentType = file.Header["Content-Type"][0]
	}

	uuidName := uuid.New().String()

	fileData, err := file.Open()
	if err != nil {
		return
	}
	defer fileData.Close()

	err = ossClient.Put(c, path.Join(defaultTmpDir, strconv.Itoa(user.ID), uuidName),
		fileData, file.Size, contentType)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, jsend.Success(uuidName))
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
