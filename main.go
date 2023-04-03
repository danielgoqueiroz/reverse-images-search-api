package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

type Result struct {
	Title     string `json:"title"`
	PageLink  string `json:"pageLink"`
	ImageLink string `json:"imageLink"`
}

func main() {

	port, host, client, notLocal, key := setup()

	app := fiber.New()
	app.Use(cors.New())
	api := app.Group("/api")
	// health check
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	// hello world
	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	// Pesquisa de imagem
	api.Get("/search", func(c *fiber.Ctx) error {
		respJSON, shouldReturn, returnValue := bingSearch(c, key, client)
		if shouldReturn {
			return returnValue
		}
		return c.SendString(string(respJSON))
	})

	if notLocal {
		host = "0.0.0.0:"
	}
	log.Fatal(app.Listen(host + port))

}

func setup() (string, string, *resty.Client, bool, string) {
	port := "5000"
	host := "localhost:"
	key := ""

	client := resty.New()
	err := godotenv.Load(".env")

	notLocal := err != nil

	if notLocal {
		log.Println("Arquivo .env não encontrado")
		port = os.Getenv("PORT")
	} else {
		log.Println("Arquivo .env encontrado")
	}

	subscriptionKey := os.Getenv("BING_SUBSCRIPTION_KEY")

	if subscriptionKey == "" {
		log.Fatal("BING_SUBSCRIPTION_KEY não configurada")
	} else {
		log.Println("BING_SUBSCRIPTION_KEY configurada", subscriptionKey)
	}

	return port, host, client, notLocal, key
}

func bingSearch(c *fiber.Ctx, key string, client *resty.Client) ([]byte, bool, error) {

	subscriptionKey := os.Getenv("BING_SUBSCRIPTION_KEY")
	if subscriptionKey == "" {
		log.Fatal("BING_SUBSCRIPTION_KEY não configurada")
	} else {
		log.Println("BING_SUBSCRIPTION_KEY configurada", subscriptionKey)
	}

	headers := map[string]string{
		"Ocp-Apim-Subscription-Key": subscriptionKey,
	}

	keyRequest := c.Request().Header.Peek("key")
	if key != string(keyRequest) {
		return nil, true, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Chave de acesso inválida"})
	}

	imageURL := c.Query("imageURL")
	if imageURL == "" {
		return nil, true, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "imageURL não informada"})
	}

	searchURL := "https://api.bing.microsoft.com/v7.0/images/visualsearch"
	queryParams := map[string]string{
		"imgurl":      imageURL,
		"mkt":         "en-us",
		"modules":     "similarimages",
		"api-version": "7.0",
	}

	resp, err := client.R().
		SetHeaders(headers).
		SetQueryParams(queryParams).
		Post(searchURL)
	if err != nil {
		return nil, true, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var results []Result
	var jsonResp map[string]interface{}
	err = json.Unmarshal(resp.Body(), &jsonResp)
	if err != nil {
		return nil, true, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	errorMsg := jsonResp["error"]
	if errorMsg != nil {
		return nil, true, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": errorMsg})
	}

	tags := jsonResp["tags"].([]interface{})
	for _, tag := range tags {
		displayName := tag.(map[string]interface{})["displayName"].(string)
		if displayName == "" {
			actions := tag.(map[string]interface{})["actions"].([]interface{})
			for _, action := range actions {
				actionType := action.(map[string]interface{})["actionType"].(string)
				if actionType == "PagesIncluding" {
					values := action.(map[string]interface{})["data"].(map[string]interface{})["value"].([]interface{})
					for _, d := range values {
						name := d.(map[string]interface{})["name"].(string)
						hostPageURL := d.(map[string]interface{})["hostPageUrl"].(string)
						contentURL := d.(map[string]interface{})["contentUrl"].(string)

						results = append(results, Result{
							Title:     name,
							PageLink:  hostPageURL,
							ImageLink: contentURL,
						})
					}
				}
			}
		}
	}

	respJSON, err := json.Marshal(results)
	if err != nil {
		return nil, true, c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return respJSON, false, nil
}
