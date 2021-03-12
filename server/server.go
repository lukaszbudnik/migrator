package server

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/graph-gophers/graphql-go"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/coordinator"
	"github.com/lukaszbudnik/migrator/data"
	"github.com/lukaszbudnik/migrator/types"
)

const (
	defaultPort     string = "8080"
	requestIDHeader string = "X-Request-ID"
)

type errorResponse struct {
	ErrorMessage string      `json:"error"`
	Details      interface{} `json:"details,omitempty"`
}

// GetPort gets the port from config or defaultPort
func GetPort(config *config.Config) string {
	if strings.TrimSpace(config.Port) == "" {
		return defaultPort
	}
	return config.Port
}

func requestIDHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.Request.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		ctx := context.WithValue(c.Request.Context(), common.RequestIDKey{}, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				common.LogPanic(c.Request.Context(), "Panic recovered: %v", err)
				if gin.IsDebugging() {
					debug.PrintStack()
				}
				c.AbortWithStatusJSON(http.StatusInternalServerError, &errorResponse{err.(string), nil})
			}
		}()
		c.Next()
	}
}

func requestLoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		common.LogInfo(c.Request.Context(), "clientIP=%v method=%v request=%v", c.ClientIP(), c.Request.Method, c.Request.URL.RequestURI())
		c.Next()
	}
}

func makeHandler(config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator, handler func(*gin.Context, *config.Config, func(context.Context, *config.Config) coordinator.Coordinator)) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c, config, newCoordinator)
	}
}

func configHandler(c *gin.Context, config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator) {
	c.YAML(200, config)
}

func schemaHandler(c *gin.Context, config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator) {
	c.String(http.StatusOK, strings.TrimSpace(data.SchemaDefinition))
}

// GraphQL endpoint
func serviceHandler(c *gin.Context, config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator) {
	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := c.ShouldBindJSON(&params); err != nil {
		common.LogError(c.Request.Context(), "Bad request: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{"Invalid request, please see documentation for valid JSON payload", nil})
		return
	}

	coordinator := newCoordinator(c.Request.Context(), config)
	defer coordinator.Dispose()
	opts := []graphql.SchemaOpt{graphql.UseFieldResolvers()}
	schema := graphql.MustParseSchema(data.SchemaDefinition, &data.RootResolver{Coordinator: coordinator}, opts...)

	response := schema.Exec(c.Request.Context(), params.Query, params.OperationName, params.Variables)
	c.JSON(http.StatusOK, response)
}

// SetupRouter setups router
func SetupRouter(versionInfo *types.VersionInfo, config *config.Config, newCoordinator func(ctx context.Context, config *config.Config) coordinator.Coordinator) *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(recovery(), requestIDHandler(), requestLoggerHandler())

	if strings.TrimSpace(config.PathPrefix) == "" {
		config.PathPrefix = "/"
	}

	r.GET(config.PathPrefix+"/", func(c *gin.Context) {
		c.JSON(http.StatusOK, versionInfo)
	})

	v2 := r.Group(config.PathPrefix + "/v2")
	v2.GET("/config", makeHandler(config, newCoordinator, configHandler))
	v2.GET("/schema", makeHandler(config, newCoordinator, schemaHandler))
	v2.POST("/service", makeHandler(config, newCoordinator, serviceHandler))

	return r
}
