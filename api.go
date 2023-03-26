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

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	client := resty.New()

	subscriptionKey := os.Getenv("BING_SUBSCRIPTION_KEY")
	if subscriptionKey == "" {
		log.Fatal("BING_SUBSCRIPTION_KEY não configurada")
	}
	headers := map[string]string{
		"Ocp-Apim-Subscription-Key": subscriptionKey,
	}

	app := fiber.New()
	app.Use(cors.New())
	api := app.Group("/api")
	// Test handler
	api.Get("/", func(c *fiber.Ctx) error {

		imageURL := c.Query("imageURL")
		if imageURL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "imageURL não informada"})
		}

		// Monta a URL da API do Bing
		searchURL := "https://api.bing.microsoft.com/v7.0/images/visualsearch"
		queryParams := map[string]string{
			"imgurl":      imageURL,
			"mkt":         "en-us",
			"modules":     "similarimages",
			"api-version": "7.0",
		}

		// Faz a solicitação de pesquisa de imagem para a API do Bing
		resp, err := client.R().
			SetHeaders(headers).
			SetQueryParams(queryParams).
			Post(searchURL)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		// Processa a resposta da API do Bing e monta a lista de resultados
		var results []Result
		var jsonResp map[string]interface{}
		err = json.Unmarshal(resp.Body(), &jsonResp)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
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

							// Adiciona o resultado à lista
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

		// Retorna a lista de resultados como JSON
		respJSON, err := json.Marshal(results)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendString(string(respJSON))
	})
	log.Fatal(app.Listen(":5000"))
}
