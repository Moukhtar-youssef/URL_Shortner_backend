package routes

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Moukhtar-youssef/URL_Shortner.git/handlers"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Create struct {
	LongURL string `param:"long_url" query:"long_url" header:"long_url" json:"long_url" xml:"long_url" form:"long_url"`
}

func SetupRoutes(DB *Storage.URLDB) *echo.Echo {
	// setting up echo server

	e := echo.New()
	// e.Use(middlewares.AllowRequests(100, 1*time.Minute, 200, 1*time.Minute))
	e.Use(middleware.CORS())
	e.Use(middleware.BodyLimit("2M"))

	// Routes
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	e.GET("/:id", func(c echo.Context) error {
		shorturl := c.Param("id")
		shorturlfull := os.Getenv("BASE_URL") + "/" + shorturl
		fmt.Println(shorturlfull)
		longurl, err := handlers.GetLongURL(DB, shorturlfull)
		if err != nil {
			log.Fatal(err)
		}

		return c.Redirect(http.StatusMovedPermanently, longurl)
	})
	e.POST("/create", func(c echo.Context) error {
		q := new(Create)
		err := c.Bind(q)
		if err != nil {
			return err
		}
		if queryVal := c.QueryParam("long_url"); queryVal != "" {
			q.LongURL = queryVal
		}
		if paramVal := c.Param("long_url"); paramVal != "" {
			q.LongURL = paramVal
		}
		if headerVal := c.Request().Header.Get("long_url"); headerVal != "" {
			q.LongURL = headerVal
		}
		var content struct {
			Response string `json:"response"`
		}
		shorturl, err := handlers.CreateShortURL(DB, q.LongURL)
		if err != nil {
			return err
		}
		content.Response = shorturl
		return c.JSON(http.StatusOK, &content)
	})

	return e
}
