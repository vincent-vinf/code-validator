package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/vincent-vinf/go-jsend"

	"github.com/vincent-vinf/code-validator/pkg/orm"
	"github.com/vincent-vinf/code-validator/pkg/perform"
	"github.com/vincent-vinf/code-validator/pkg/util"
	"github.com/vincent-vinf/code-validator/pkg/util/config"
	"github.com/vincent-vinf/code-validator/pkg/util/db"
	"github.com/vincent-vinf/code-validator/pkg/util/jwtx"
	"github.com/vincent-vinf/code-validator/pkg/util/mq"
	"github.com/vincent-vinf/code-validator/pkg/util/oss"
	"github.com/vincent-vinf/code-validator/pkg/util/zip"
)

const (
	defaultTmpDir         = "tmp"
	defaultCaseFileName   = "case.zip"
	defaultUnzipDir       = "unzip-out"
	defaultCaseInFileExt  = ".in"
	defaultCaseOutFileExt = ".out"
	defaultCaseDir        = "cases"

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
	router.POST("/case/file", uploadCaseFile)
	router.POST("/case", uploadCase)

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
	uuidName, err := uploadFileToOSS(c, file, user.ID)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, jsend.Success(uuidName))
}

func uploadCaseFile(c *gin.Context) {
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
	tempDir, err := os.MkdirTemp("", "case")
	if err != nil {
		return
	}
	defer os.RemoveAll(tempDir)
	caseFilePath := path.Join(tempDir, defaultCaseFileName)
	err = c.SaveUploadedFile(file, caseFilePath)
	if err != nil {
		return
	}
	unzipDir := path.Join(tempDir, defaultUnzipDir)
	err = zip.UnzipSource(caseFilePath, unzipDir)
	if err != nil {
		return
	}
	uuidName := c.Query("uuid")
	uuidName, err = newOrCheckUUID(uuidName)
	if err != nil {
		return
	}

	res, err := putCasesFromDir(c, unzipDir, path.Join(getUserTempDir(user.ID), defaultCaseDir, uuidName))
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, jsend.Success(map[string]any{
		"cases": res,
		"uuid":  uuidName,
	}))
}

func uploadCase(c *gin.Context) {
	t, _ := c.Get(jwtx.IdentityKey)
	user := t.(*jwtx.TokenUserInfo)

	var err error
	defer func() {
		if err != nil {
			c.JSON(http.StatusInternalServerError, jsend.SimpleErr(err.Error()))
		}
	}()

	uuidName := c.Query("uuid")
	uuidName, err = newOrCheckUUID(uuidName)
	if err != nil {
		return
	}
	cases := make([]Case, 0)
	if err = c.BindJSON(&cases); err != nil {
		return
	}

	ossDir := path.Join(getUserTempDir(user.ID), defaultCaseDir, uuidName)
	res, err := putCases(c, cases, ossDir)
	if err != nil {
		return
	}
	c.JSON(http.StatusOK, jsend.Success(map[string]any{
		"cases": res,
		"uuid":  uuidName,
	}))
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

func uploadFileToOSS(ctx context.Context, file *multipart.FileHeader, uid int) (name string, err error) {
	log.Info("filename:", file.Filename)

	contentType := defaultContentType
	if len(file.Header["Content-Type"]) > 0 {
		contentType = file.Header["Content-Type"][0]
	}

	uuidName := uuid.New().String()

	fileData, err := file.Open()
	if err != nil {
		return "", err
	}
	defer fileData.Close()

	err = ossClient.Put(ctx, path.Join(getUserTempDir(uid), uuidName),
		fileData, file.Size, contentType)
	if err != nil {
		return "", err
	}

	return uuidName, nil
}

func getUserTempDir(uid int) string {
	return path.Join(defaultTmpDir, strconv.Itoa(uid))
}

func putCasesFromDir(ctx context.Context, dir, ossDir string) (res []perform.TestCase, err error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	if len(files) == 1 && files[0].IsDir() {
		dir = path.Join(dir, files[0].Name())
		files, err = os.ReadDir(dir)
		if err != nil {
			return
		}
	}

	outNameMap := make(map[string]struct{})
	for i := range files {
		if files[i].IsDir() {
			continue
		}
		name := files[i].Name()
		if ext := path.Ext(name); ext == defaultCaseOutFileExt {
			outNameMap[name] = struct{}{}
		}
	}

	for i := range files {
		if files[i].IsDir() {
			// ignore dir
			continue
		}
		name := files[i].Name()
		if ext := path.Ext(name); ext == defaultCaseInFileExt {
			caseName := name[:len(name)-len(ext)]
			outName := caseName + defaultCaseOutFileExt
			if _, ok := outNameMap[outName]; !ok {
				continue
			}
			ossInPath := path.Join(ossDir, name)
			ossOutPath := path.Join(ossDir, outName)

			if err = putLocalTextFile(ctx, path.Join(dir, name), ossInPath); err != nil {
				return
			}
			if err = putLocalTextFile(ctx, path.Join(dir, outName), ossOutPath); err != nil {
				return
			}
			t := perform.TestCase{
				Name: caseName,
				In: perform.File{
					OssPath: ossInPath,
				},
				Out: perform.File{
					OssPath: ossOutPath,
				},
			}

			res = append(res, t)
		}
	}
	return
}

func putCases(ctx context.Context, cases []Case, ossDir string) (res []perform.TestCase, err error) {
	for i := range cases {
		ossInPath := path.Join(ossDir, cases[i].Name+defaultCaseInFileExt)
		ossOutPath := path.Join(ossDir, cases[i].Name+defaultCaseOutFileExt)

		if err = putTextFile(ctx, []byte(cases[i].In), ossInPath); err != nil {
			return
		}
		if err = putTextFile(ctx, []byte(cases[i].Out), ossOutPath); err != nil {
			return
		}

		t := perform.TestCase{
			Name: cases[i].Name,
			In: perform.File{
				OssPath: ossInPath,
			},
			Out: perform.File{
				OssPath: ossOutPath,
			},
		}

		res = append(res, t)
	}

	return
}

func putLocalTextFile(ctx context.Context, path, ossPath string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	fstat, err := file.Stat()
	if err != nil {
		return err
	}
	return ossClient.Put(ctx, ossPath, file, fstat.Size(), defaultContentType)
}

func putTextFile(ctx context.Context, data []byte, ossPath string) error {
	return ossClient.Put(ctx, ossPath, bytes.NewReader(data), int64(len(data)), defaultContentType)
}

func newOrCheckUUID(uuidName string) (string, error) {
	if uuidName == "" {
		uuidName = uuid.New().String()
	} else if _, err := uuid.Parse(uuidName); err != nil {
		return "", fmt.Errorf("invalid uuid, err: %w", err)
	}

	return uuidName, nil
}

type Case struct {
	Name string
	In   string
	Out  string
}
