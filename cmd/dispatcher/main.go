package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"

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
	defaultBatchDir       = "batch"
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

	// todo: for test
	perform.SetOssClient(ossClient)

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
	var err error
	defer func() {
		if err != nil {
			c.JSON(http.StatusInternalServerError, jsend.SimpleErr(err.Error()))
		}
	}()
	type Request struct {
		Name          string
		Verifications []*perform.Verification
	}
	req := &Request{}
	if err = c.BindJSON(req); err != nil {
		return
	}

	var vfs []*orm.Verification
	for _, vf := range req.Verifications {
		var data []byte
		data, err = json.Marshal(vf)
		if err != nil {
			return
		}
		vfs = append(vfs, &orm.Verification{
			Name: vf.Name,
			Data: data,
		})
	}
	t, _ := c.Get(jwtx.IdentityKey)
	user := t.(*jwtx.TokenUserInfo)
	batch := &orm.Batch{
		Name:          req.Name,
		UserID:        user.ID,
		CreatedAt:     time.Now(),
		Verifications: vfs,
	}
	log.Info(batch)
	if err = db.AddBatch(batch); err != nil {
		return
	}
	for _, vf := range req.Verifications {
		err = moveRefFile(c, user.ID, batch.ID, vf)
		if err != nil {
			return
		}
	}

	// test
	var res []*perform.Report
	for _, vf := range req.Verifications {
		log.Info(vf.String())
		var rep *perform.Report
		rep, err = perform.Perform(vf, "t/in2out.py")
		if err != nil {
			return
		}
		res = append(res, rep)
	}

	c.JSON(http.StatusOK, jsend.Success(res))
}

func moveRefFile(ctx context.Context, uid, batchID int, vf *perform.Verification) error {
	userTempDir := getUserTempDir(uid)
	batchDir := getBatchDir(batchID)
	if vf.Code != nil {
		files, err := moveOssFiles(ctx, vf.Code.Files, userTempDir, batchDir)
		if err != nil {
			return err
		}
		vf.Code.Files = files
		files, err = moveOssFiles(ctx, vf.Code.Init.Files, userTempDir, batchDir)
		if err != nil {
			return err
		}
		vf.Code.Init.Files = files
		for i := range vf.Code.Cases {
			files, err = moveOssFiles(ctx, []perform.File{vf.Code.Cases[i].In}, userTempDir, batchDir)
			if err != nil {
				return err
			}
			vf.Code.Cases[i].In = files[0]
			files, err = moveOssFiles(ctx, []perform.File{vf.Code.Cases[i].Out}, userTempDir, batchDir)
			if err != nil {
				return err
			}
			vf.Code.Cases[i].Out = files[0]
		}
	} else if vf.Custom != nil {
		files, err := moveOssFiles(ctx, vf.Custom.Files, userTempDir, batchDir)
		if err != nil {
			return err
		}
		vf.Custom.Files = files
	}

	return nil
}
func moveOssFiles(ctx context.Context, files []perform.File, src, dst string) ([]perform.File, error) {
	var res []perform.File
	for _, f := range files {
		d := path.Join(dst, f.OssPath)
		err := ossClient.Move(ctx, path.Join(src, f.OssPath), d)
		if err != nil {
			return nil, err
		}
		f.OssPath = d
		res = append(res, f)
	}

	return res, nil
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

	res, err := putCasesFromDir(c, unzipDir, user.ID, path.Join(defaultCaseDir, uuidName))
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
	res, err := putCases(c, cases, user.ID, path.Join(defaultCaseDir, uuidName))
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

func getBatchDir(batchID int) string {
	return path.Join(defaultBatchDir, strconv.Itoa(batchID))
}

func putCasesFromDir(ctx context.Context, dir string, uid int, ossDir string) (res []perform.TestCase, err error) {
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

	userTempDir := getUserTempDir(uid)
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

			if err = putLocalTextFile(ctx, path.Join(dir, name), path.Join(userTempDir, ossInPath)); err != nil {
				return
			}
			if err = putLocalTextFile(ctx, path.Join(dir, outName), path.Join(userTempDir, ossOutPath)); err != nil {
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

func putCases(ctx context.Context, cases []Case, uid int, ossDir string) (res []perform.TestCase, err error) {
	userTempDir := getUserTempDir(uid)

	for i := range cases {
		ossInPath := path.Join(ossDir, cases[i].Name+defaultCaseInFileExt)
		ossOutPath := path.Join(ossDir, cases[i].Name+defaultCaseOutFileExt)

		if err = putTextFile(ctx, []byte(cases[i].Input), path.Join(userTempDir, ossInPath)); err != nil {
			return
		}
		if err = putTextFile(ctx, []byte(cases[i].Output), path.Join(userTempDir, ossOutPath)); err != nil {
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
	Name   string
	Input  string
	Output string
}
