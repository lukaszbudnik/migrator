package server

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/graph-gophers/graphql-go"
	"gopkg.in/go-playground/validator.v9"

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

type migrationsPostRequest struct {
	Response types.MigrationsResponseType `json:"response" binding:"required,response"`
	Mode     types.MigrationsModeType     `json:"mode" binding:"required,mode"`
}

type tenantsPostRequest struct {
	Name string `json:"name" binding:"required"`
	migrationsPostRequest
}

type migrationsSuccessResponse struct {
	Results           *types.MigrationResults `json:"results"`
	AppliedMigrations []types.Migration       `json:"appliedMigrations,omitempty"`
}

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
		if strings.Contains(c.Request.URL.Path, "/v1/") {
			c.Header("Deprecation", `version="v2020.1.0"`)
			c.Header("Link", `<https://github.com/lukaszbudnik/migrator/#v2---graphql-based-api>; rel="successor-version"`)
		}
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

func migrationsSourceHandler(c *gin.Context, config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator) {
	coordinator := newCoordinator(c.Request.Context(), config)
	defer coordinator.Dispose()
	migrations := coordinator.GetSourceMigrations(nil)
	common.LogInfo(c.Request.Context(), "Returning source migrations: %v", len(migrations))
	c.JSON(http.StatusOK, migrations)
}

func migrationsAppliedHandler(c *gin.Context, config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator) {
	coordinator := newCoordinator(c.Request.Context(), config)
	defer coordinator.Dispose()
	dbMigrations := coordinator.GetAppliedMigrations()
	common.LogInfo(c.Request.Context(), "Returning applied migrations: %v", len(dbMigrations))
	c.JSON(http.StatusOK, dbMigrations)
}

func migrationsPostHandler(c *gin.Context, config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator) {
	var request migrationsPostRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		common.LogError(c.Request.Context(), "Error reading request: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{"Invalid request, please see documentation for valid JSON payload", nil})
		return
	}

	coordinator := newCoordinator(c.Request.Context(), config)
	defer coordinator.Dispose()

	if ok, offendingMigrations := coordinator.VerifySourceMigrationsCheckSums(); !ok {
		common.LogError(c.Request.Context(), "Checksum verification failed for migrations: %v", len(offendingMigrations))
		c.AbortWithStatusJSON(http.StatusFailedDependency, errorResponse{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		return
	}

	results, appliedMigrations := coordinator.ApplyMigrations(request.Mode)

	common.LogInfo(c.Request.Context(), "Returning applied migrations: %v", len(appliedMigrations))

	var response *migrationsSuccessResponse
	switch request.Response {
	case types.ResponseTypeFull:
		response = &migrationsSuccessResponse{results, appliedMigrations}
	case types.ResponseTypeList:
		for i := range appliedMigrations {
			appliedMigrations[i].Contents = ""
		}
		response = &migrationsSuccessResponse{results, appliedMigrations}
	default:
		response = &migrationsSuccessResponse{results, nil}
	}

	c.JSON(http.StatusOK, response)
}

func tenantsGetHandler(c *gin.Context, config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator) {
	coordinator := newCoordinator(c.Request.Context(), config)
	defer coordinator.Dispose()
	// starting v2019.1.0 GetTenants returns a slice of Tenant struct
	// /v1 API returns a slice of strings and we must maintain backward compatibility
	tenants := coordinator.GetTenants()
	tenantNames := []string{}
	for _, t := range tenants {
		tenantNames = append(tenantNames, t.Name)
	}
	common.LogInfo(c.Request.Context(), "Returning tenants: %v", len(tenants))
	c.JSON(http.StatusOK, tenantNames)
}

func tenantsPostHandler(c *gin.Context, config *config.Config, newCoordinator func(context.Context, *config.Config) coordinator.Coordinator) {
	var request tenantsPostRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		common.LogError(c.Request.Context(), "Bad request: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{"Invalid request, please see documentation for valid JSON payload", nil})
		return
	}

	coordinator := newCoordinator(c.Request.Context(), config)
	defer coordinator.Dispose()

	if ok, offendingMigrations := coordinator.VerifySourceMigrationsCheckSums(); !ok {
		common.LogError(c.Request.Context(), "Checksum verification failed for migrations: %v", len(offendingMigrations))
		c.AbortWithStatusJSON(http.StatusFailedDependency, errorResponse{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		return
	}

	results, appliedMigrations := coordinator.AddTenantAndApplyMigrations(request.Mode, request.Name)

	common.LogInfo(c.Request.Context(), "Tenant %v added, migrations applied: %v", request.Name, len(appliedMigrations))

	var response *migrationsSuccessResponse
	switch request.Response {
	case types.ResponseTypeFull:
		response = &migrationsSuccessResponse{results, appliedMigrations}
	case types.ResponseTypeList:
		for i := range appliedMigrations {
			appliedMigrations[i].Contents = ""
		}
		response = &migrationsSuccessResponse{results, appliedMigrations}
	default:
		response = &migrationsSuccessResponse{results, nil}
	}

	c.JSON(http.StatusOK, response)
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

	// there is something seriously wrong with validator and its gin integration
	binding.Validator = new(defaultValidator)
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("response", types.ValidateMigrationsResponseType)
		v.RegisterValidation("mode", types.ValidateMigrationsModeType)
	}

	if strings.TrimSpace(config.PathPrefix) == "" {
		config.PathPrefix = "/"
	}

	r.GET(config.PathPrefix+"/", func(c *gin.Context) {
		c.JSON(http.StatusOK, versionInfo)
	})

	v1 := r.Group(config.PathPrefix + "/v1")

	v1.GET("/config", makeHandler(config, newCoordinator, configHandler))

	v1.GET("/tenants", makeHandler(config, newCoordinator, tenantsGetHandler))
	v1.POST("/tenants", makeHandler(config, newCoordinator, tenantsPostHandler))

	v1.GET("/migrations/source", makeHandler(config, newCoordinator, migrationsSourceHandler))
	v1.GET("/migrations/applied", makeHandler(config, newCoordinator, migrationsAppliedHandler))
	v1.POST("/migrations", makeHandler(config, newCoordinator, migrationsPostHandler))

	v2 := r.Group(config.PathPrefix + "/v2")
	v2.GET("/config", makeHandler(config, newCoordinator, configHandler))
	v2.GET("/schema", makeHandler(config, newCoordinator, schemaHandler))
	v2.POST("/service", makeHandler(config, newCoordinator, serviceHandler))

	return r
}
