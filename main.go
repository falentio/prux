package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			fmt.Fprintf(os.Stderr, "error: %v", err)
			return err
		},
	})
	app.All("/tripay/*", func (c *fiber.Ctx) error {
		return proxy(c, "https://tripay.co.id/api/")
	})
	app.All("/tripay-sandbox/*", func (c *fiber.Ctx) error {
		return proxy(c, "https://tripay.co.id/api-sandbox/")
	})
	app.All("/digiflazz/*", func (c *fiber.Ctx) error {
		return proxy(c, "https://api.digiflazz.com/v1/")
	})

	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		panic(err)
	}
}

func proxy(c *fiber.Ctx, endpoint string) error {
	u, err := url.Parse(endpoint)
	if err != nil {
		return err
	}
	u, err = u.Parse(c.Params("*"))
	if err != nil {
		return err
	}
	r, err := url.Parse(c.OriginalURL())
	if err != nil {
		return err
	}
	u.RawQuery = r.Query().Encode()
	body := bytes.NewBuffer(c.Body())
	req, err := http.NewRequest(c.Method(), u.String(), body)
	if err != nil {
		return err
	}
	for k, v := range c.GetReqHeaders() {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for k, v := range resp.Header {
		c.Append(k, v...)
	}
	fmt.Println(resp.StatusCode, u.String(), c.OriginalURL(), c.Method())
	_, err = io.Copy(c, resp.Body)
	return err
}