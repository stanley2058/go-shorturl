package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/joho/godotenv/autoload"

	db "github.com/stanley2058/shorturl/redis"
	structure "github.com/stanley2058/shorturl/structure"
)

var preservedWords = map[string]bool{"all": true}

func main() {
	app := fiber.New()
	app.Use(cors.New())

	app.Static("/", "./static")
	app.Get("/all", getAllEntries)
	app.Post("/", createShortUrl)
	app.Get("/:shorten", redirect)
	app.Post("/:shorten", updateShortenUrl)
	app.Delete("/:shorten", deleteShortenUrl)

	app.Listen(fmt.Sprintf(":%s", os.Getenv("PORT")))
}

func redirect(c *fiber.Ctx) error {
	shortenKey, _ := url.QueryUnescape(c.Params("shorten"))
	res, err := db.Get(shortenKey)
	if err != nil {
		return handleStatusCode(c, fiber.StatusNotFound, err)
	}
	var body structure.UrlObject
	err = json.Unmarshal([]byte(res), &body)
	if err != nil {
		return handleServerError(c, err)
	}
	return c.Redirect(body.Url)
}

func createShortUrl(c *fiber.Ctx) error {
	var body structure.CreateRecordContext
	if err := c.BodyParser(&body); err != nil {
		return handleStatusCode(c, fiber.StatusBadRequest)
	}

	// url cannot be empty
	if body.Url == "" {
		return handleStatusCode(c, fiber.StatusBadRequest)
	}

	key, _ := url.QueryUnescape(body.Shorten)
	// shorten key cannot preserved word
	if _, ok := preservedWords[key]; ok {
		return handleStatusCode(c, fiber.StatusBadRequest)
	}
	// generate shorten key if not provided
	if key == "" {
		key = db.GenerateRandomKey()
	} else if val, err := db.Get(key); err == nil {
		// check if shorten key already exists
		var bodyOld structure.UrlObject
		err = json.Unmarshal([]byte(val), &bodyOld)
		if err != nil {
			return handleServerError(c, err)
		}
		if bodyOld.Activated {
			return handleStatusCode(c, fiber.StatusBadRequest)
		}
	}

	// save shorten key
	err := db.Save(key, body.Url)
	if err != nil {
		return handleServerError(c, err)
	}

	val, _ := json.Marshal(structure.UrlObject{
		Url:       body.Url,
		Activated: true,
	})
	db.Save(key, string(val))

	return c.Status(fiber.StatusCreated).JSON(struct {
		Shorten string `json:"shorten"`
	}{
		Shorten: key,
	})
}

func getAllEntries(c *fiber.Ctx) error {
	res, err := db.GetAllEntries()
	if err != nil {
		return handleServerError(c, err)
	}
	return c.JSON(res)
}

func updateShortenUrl(c *fiber.Ctx) error {
	shortenKey := c.Params("shorten")
	var body structure.UpdateRecordContext
	if err := c.BodyParser(&body); err != nil {
		return handleStatusCode(c, fiber.StatusBadRequest)
	}
	// new shorten key cannot be empty
	newShorten, _ := url.QueryUnescape(body.NewShorten)
	if newShorten == "" {
		newShorten = db.GenerateRandomKey()
	}
	// new shorten key cannot be preserved word
	if _, ok := preservedWords[newShorten]; ok {
		return handleStatusCode(c, fiber.StatusBadRequest)
	}

	res, err := db.Get(shortenKey)
	if err != nil {
		return handleStatusCode(c, fiber.StatusNotFound, err)
	}
	var bodyOld structure.UrlObject
	if err = json.Unmarshal([]byte(res), &bodyOld); err != nil {
		return handleServerError(c, err)
	}

	// set old key to inactive
	bodyOld.Activated = false
	oldVal, _ := json.Marshal(bodyOld)
	err = db.Save(shortenKey, string(oldVal))
	if err != nil {
		return handleServerError(c, err)
	}

	// save new key
	val, _ := json.Marshal(structure.UrlObject{
		Url:       bodyOld.Url,
		Activated: true,
	})
	err = db.Save(newShorten, string(val))
	if err != nil {
		return handleServerError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(struct {
		Shorten string `json:"shorten"`
	}{
		Shorten: newShorten,
	})
}

func deleteShortenUrl(c *fiber.Ctx) error {
	shortenKey := c.Params("shorten")
	if err := db.Delete(shortenKey); err != nil {
		return handleServerError(c, err)
	}
	return handleStatusCode(c, fiber.StatusOK)
}

func handleServerError(c *fiber.Ctx, err error) error {
	return handleStatusCode(c, fiber.StatusInternalServerError, err)
}
func handleStatusCode(c *fiber.Ctx, statusCode int, err ...error) error {
	if len(err) > 0 && err[0] != nil {
		log.Println(err[0])
	}
	return c.SendStatus(statusCode)
}
