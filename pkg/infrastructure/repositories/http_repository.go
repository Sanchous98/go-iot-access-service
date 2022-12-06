package repositories

import (
	"bitbucket.org/4suites/iot-service-golang/pkg/application"
	"bitbucket.org/4suites/iot-service-golang/pkg/domain/logger"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"strings"
)

type meta struct {
	//Locale     string `json:"locale"`
	//Version    int    `json:"version"`
	Pagination struct {
		Total int `json:"total"`
		//Count       int `json:"count"`
		Limit       int `json:"limit"`
		CurrentPage int `json:"currentPage"`
		TotalPages  int `json:"totalPages"`
	} `json:"pagination"`
}

type responseFindShape[T application.WithResource] struct {
	Data T    `json:"data"`
	Meta meta `json:"meta"`
}

type responseFindAllShape[T application.WithResource] struct {
	Data []T  `json:"data"`
	Meta meta `json:"meta"`
}

type RegistryRepository[T application.WithResource] struct {
	ApiBaseUrl string `env:"REGISTRY_API_URL"`
	ApiKey     string `env:"REGISTRY_API_KEY"`

	log logger.Logger `inject:""`

	client *fiber.Client
}

func (r *RegistryRepository[T]) Constructor() {
	r.client = fiber.AcquireClient()
}

func (r *RegistryRepository[T]) Destructor() {
	fiber.ReleaseClient(r.client)
	r.client = nil
}

func (r *RegistryRepository[T]) getUrl() string {
	var model application.WithResource = *new(T)
	return r.ApiBaseUrl + "/" + strings.TrimPrefix(model.GetResource(), "/")
}

func (r *RegistryRepository[T]) find(id uuid.UUID, params map[string]any) T {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s/%s?key=%s", r.getUrl(), id.String(), r.ApiKey))

	if params != nil {
		for key, value := range params {
			builder.WriteString(fmt.Sprintf("&%s=%v", key, value))
		}
	}

	var responseBody responseFindShape[T]

	agent := r.client.Get(builder.String()).Add(fiber.HeaderAccept, fiber.MIMEApplicationJSON)
	code, _, errors := agent.Struct(&responseBody)

	if len(errors) != 0 {
		r.log.Errorln(errors)
		return *new(T)
	}

	if code >= 400 {
		r.log.Debugf("Request failed with HTTP code: %d\n, URL: %s", code, agent.Request().String())
		return *new(T)
	}

	return responseBody.Data
}

func (r *RegistryRepository[T]) findAll(condition map[string]any) []T {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s?key=%s", r.getUrl(), r.ApiKey))

	if condition != nil {
		for key, value := range condition {
			builder.WriteString(fmt.Sprintf("&%s=%v", key, value))
		}
	}

	var responseBody responseFindAllShape[T]

	query := builder.String()
	agent := r.client.Get(query).Add(fiber.HeaderAccept, fiber.MIMEApplicationJSON)
	code, _, errors := agent.Struct(&responseBody)

	if code >= 400 {
		r.log.Debugf("Request failed with HTTP code: %d\n, URL: %s", code, agent.Request().URI().String())
		return nil
	}

	if len(errors) != 0 {
		r.log.Errorln(errors)
		return nil
	}

	if responseBody.Meta.Pagination.Total <= responseBody.Meta.Pagination.Limit {
		return responseBody.Data
	}

	result := make([]T, 0, responseBody.Meta.Pagination.Total)
	result = append(result, responseBody.Data...)

	for i := responseBody.Meta.Pagination.CurrentPage + 1; i <= responseBody.Meta.Pagination.TotalPages; i++ {
		agent = r.client.Get(query+fmt.Sprintf("&page=%d", i)).Add(fiber.HeaderAccept, fiber.MIMEApplicationJSON)
		agent.Struct(&responseBody)
		result = append(result, responseBody.Data...)
	}

	return result
}
