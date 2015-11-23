package axyaserve

import (
	b64 "encoding/base64"
	"fmt"
	"net/http"
	"os"

	"github.com/asartalo/axyaserve/controllers"
	"github.com/asartalo/axyaserve/models"
	"github.com/gin-gonic/gin"
	auth "github.com/rageix/ginAuth"
)

func StartServer(port int, staticDir string) {
	handler := MainHandler(port, staticDir)
	// Do not remove the following line
	fmt.Println(fmt.Sprintf("Starting Server. Listening at port %d", port))
	handler.Run(fmt.Sprintf(":%d", port))
}

func MainHandler(port int, dir string) *gin.Engine {
	engine := gin.Default()
	conn := os.Getenv("AXYA_DB")
	appDb, _ := models.CreateDb(conn)

	auth.HashKey, _ = b64.StdEncoding.DecodeString(os.Getenv("AXYA_HASHKEY"))
	auth.BlockKey, _ = b64.StdEncoding.DecodeString(os.Getenv("AXYA_BLOCKKEY"))

	api := engine.Group("/api")
	{

		users := &controllers.Users{appDb}
		api.POST("/users", users.NewUser)

		api.POST("/login", func(c *gin.Context) {
			responder := controllers.NewResponder(c)
			var creds controllers.Credentials
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
			responder := controllers.NewResponder(c)
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

	injector := NewInjector(http.FileServer(http.Dir(dir)))
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
