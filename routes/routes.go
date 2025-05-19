package routes

import (
	"log"
	"net/http"

	"github.com/Moukhtar-youssef/URL_Shortner.git/handlers"
	Storage "github.com/Moukhtar-youssef/URL_Shortner.git/storage"
	"github.com/labstack/echo/v4"
)

func SetupRoutes(DB *Storage.URLDB, e *echo.Echo) {

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
		long_url := c.QueryParam("long_url")
		var content struct {
			Response string `json:"response"`
		}
		shorturl, err := handlers.CreateShortURL(DB, long_url)
		if err != nil {
			return err
		}
		content.Response = shorturl
		return c.JSON(http.StatusOK, &content)
	})
}
