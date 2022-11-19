//go:build python

package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/vincent-vinf/code-validator/pkg/runner"
	"github.com/vincent-vinf/code-validator/pkg/util/dispatcher"
)

var (
	idRing *dispatcher.Dispatcher
)

func init() {
	var err error
	idRing, err = dispatcher.NewDispatcher(800, 990)
	if err != nil {
		panic(err)
	}
}

func main() {
	r := gin.New()
	gin.SetMode(gin.ReleaseMode)
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(Cors())
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"message": "Page not found"})
	})
	r.POST("/python", runPython)

	if err := r.Run(":9000"); err != nil {
		log.Fatalln(err)
	}
}
func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

type RunRequest struct {
	Code  string
	Input string
}

type RunResponse struct {
	Pass   bool   `json:"pass"`
	Output string `json:"output"`
	Error  string `json:"error,omitempty"`
}

func runPython(c *gin.Context) {
	req := &RunRequest{}
	err := c.Bind(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

		return
	}
	id, err := idRing.Get()
	if err != nil {
		if errors.Is(err, dispatcher.SpaceFullErr) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"message": "the server is busy, please try again later",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}
	res := &RunResponse{
		Pass: true,
	}
	output, err := runner.Run(id, []byte(req.Input), []byte(req.Code))
	if err != nil {
		res.Error = err.Error()
		res.Pass = false
	}
	res.Output = string(output)

	c.JSON(http.StatusOK, res)

	return
}
