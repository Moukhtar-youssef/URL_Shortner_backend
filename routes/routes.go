package routes

import (
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Moukhtar-youssef/URL_Shortner.git/handlers"
	"github.com/Moukhtar-youssef/URL_Shortner.git/middlewares"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Create struct {
	LongURL string `param:"long_url" query:"long_url" header:"long_url" json:"long_url" xml:"long_url" form:"long_url"`
}

func SetupRoutes(DB *Storage.URLDB) *echo.Echo {
	e := echo.New()
	c := jaegertracing.New(e, nil)
	defer func(c io.Closer) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)
	e.Use(middlewares.AllowRequests(100, 1*time.Minute, 200, 1*time.Minute))
	e.Use(middleware.CORS())
	e.Use(middleware.BodyLimit("2M"))

	// Routes

	e.GET("/:id", func(c echo.Context) error {
		req := c.Request()
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		host := req.Host
		baseURL := scheme + "://" + host
		shorturl := c.Param("id")
		shorturlfull := baseURL + "/" + shorturl
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
