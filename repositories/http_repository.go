package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/utils"
	"fmt"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"log"
	"strings"
)

type WithEndpoint interface {
	GetEndpoint() string
}

type Repository[T WithEndpoint] interface {
	Find(id utils.UUID) T
	FindAll() []T
}

type RegistryRepository[T WithEndpoint] struct {
	ApiBaseUrl string `env:"REGISTRY_API_URL"`
	ApiKey     string `env:"REGISTRY_API_KEY"`
	client     *fiber.Client
}

func (r *RegistryRepository[T]) Constructor() {
	r.client = fiber.AcquireClient()
}

func (r *RegistryRepository[T]) Destructor() {
	fiber.ReleaseClient(r.client)
	r.client = nil
}

func (r *RegistryRepository[T]) GetUrl() string {
	var model WithEndpoint = *new(T)
	return r.ApiBaseUrl + model.GetEndpoint()
}

func (r *RegistryRepository[T]) Find(id utils.UUID) (result T) {
	return r.find(id)
}

func (r *RegistryRepository[T]) find(id utils.UUID) T {
	agent := r.client.Get(fmt.Sprintf("%s/%s?key=%s", r.GetUrl(), id.String(), r.ApiKey)).Add("Accept", "application/json")
	code, body, errors := agent.Bytes()

	if len(errors) != 0 {
		log.Println(errors)
	}

	if code >= 400 {
		log.Printf("Request failed with HTTP code: %d\n, URL: %s", code, agent.Request().String())
	}

	responseBody := struct {
		Data T `json:"data"`
	}{}
	_ = json.Unmarshal(body, &responseBody)

	return responseBody.Data
}

func (r *RegistryRepository[T]) FindAll() []T {
	return r.findAll(map[string]any{
		"enabled": 1,
		"claimed": 1,
	})
}

func (r *RegistryRepository[T]) findAll(condition map[string]any) []T {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s?key=%s", r.GetUrl(), r.ApiKey))

	if condition != nil {
		for key, value := range condition {
			builder.WriteString(fmt.Sprintf("&%s=%v", key, value))
		}
	}

	agent := r.client.Get(builder.String()).Add("Accept", "application/json")
	code, body, errors := agent.Bytes()

	if len(errors) != 0 {
		log.Println(errors)
	}

	if code >= 400 {
		log.Printf("Request failed with HTTP code: %d\n, URL: %s", code, agent.Request().String())
	}

	responseBody := struct {
		Data []T `json:"data"`
	}{}
	_ = json.Unmarshal(body, &responseBody)

	return responseBody.Data
}
