package axyaserve

import (
	b64 "encoding/base64"
	"fmt"
	"github.com/asartalo/axyaserve/model"
	"github.com/gin-gonic/gin"
	auth "github.com/rageix/ginAuth"
	"net/http"
	"os"
)

type Credentials struct {
	Name     string `form:"name" json:"name" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func StartServer(port int, staticDir string) {
	handler := MainHandler(port, staticDir)
	// Do not remove the following line
	fmt.Println(fmt.Sprintf("Starting Server. Listening at port %d", port))
	handler.Run(fmt.Sprintf(":%d", port))
}

type Responder struct {
	context *gin.Context
}

func NewResponder(c *gin.Context) *Responder {
	return &Responder{c}
}

func (j *Responder) Error(status int, message, details string) {
	j.context.JSON(status, gin.H{"message": message, "details": details})
}

func (j *Responder) Okay(message string, payload gin.H) {
	j.context.JSON(http.StatusOK, gin.H{"message": message, "payload": payload})
}

func (j *Responder) Created(message string, payload gin.H) {
	j.context.JSON(http.StatusCreated, gin.H{"message": message, "payload": payload})
}

func MainHandler(port int, dir string) *gin.Engine {
	engine := gin.Default()
	conn := os.Getenv("AXYA_DB")
	appDb, _ := model.CreateDb(conn)

	auth.HashKey, _ = b64.StdEncoding.DecodeString(os.Getenv("AXYA_HASHKEY"))
	auth.BlockKey, _ = b64.StdEncoding.DecodeString(os.Getenv("AXYA_BLOCKKEY"))

	api := engine.Group("/api")
	{
		api.POST("/users", func(c *gin.Context) {
			responder := NewResponder(c)
			var creds Credentials
			err := c.BindJSON(&creds)
			if err != nil {
				responder.Error(
					http.StatusBadRequest,
					"Make sure credentials are complete.",
					err.Error(),
				)
				return
			}
			user, err := appDb.NewUser(creds.Name, creds.Password)
			if err != nil {
				responder.Error(
					http.StatusInternalServerError,
					"Error: "+err.Error(),
					err.Error(),
				)
				return
			}
			responder.Created("User created", gin.H{"name": user.Name})
		})

		api.POST("/login", func(c *gin.Context) {
			responder := NewResponder(c)
			var creds Credentials
			err := c.BindJSON(&creds)
			if err != nil {
				responder.Error(
					http.StatusBadRequest,
					"Make sure credentials are complete.",
					err.Error(),
				)
				return
			}

			user, err := appDb.Authenticate(creds.Name, creds.Password)
			if err != nil {
				responder.Error(
					http.StatusUnauthorized,
					"Authentication error",
					err.Error(),
				)
				return
			}
			err = auth.Login(c, map[string]string{"name": user.Name})
			if err != nil {
				responder.Error(
					http.StatusUnauthorized,
					"Authentication error",
					err.Error(),
				)
				return
			}
			responder.Okay("Login successful", gin.H{"name": user.Name})
		})

		api.GET("/logout", func(c *gin.Context) {
			auth.Logout(c)
		})

		auth.Unauthorized = func(c *gin.Context) {
			responder := NewResponder(c)
			responder.Error(
				http.StatusUnauthorized,
				"Authentication error",
				"You are not logged in",
			)
			c.Abort()
		}

		authenticate := api.Group("/")
		authenticate.Use(auth.Use)
	}

	injector := Injector(http.FileServer(http.Dir(dir)))
	injector.Inject("text/html", InjectLiveReload)
	wrapedInjector := gin.WrapH(injector)
	engine.GET("/components/*filepath", wrapedInjector)
	engine.GET("/css/*filepath", wrapedInjector)
	engine.GET("/js/*filepath", wrapedInjector)
	engine.GET("/style-guide/*filepath", wrapedInjector)
	engine.GET("/templates/*filepath", wrapedInjector)
	engine.GET("/", wrapedInjector)

	return engine
}
