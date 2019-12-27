package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/lukaszbudnik/migrator/db"
	"github.com/lukaszbudnik/migrator/loader"
	"github.com/lukaszbudnik/migrator/migrations"
	"github.com/lukaszbudnik/migrator/notifications"
	"github.com/lukaszbudnik/migrator/types"
)

const (
	defaultPort     string = "8080"
	requestIDHeader string = "X-Request-Id"
)

type migrationsResponseType string
type migrationsModeType string

const (
	responseTypeSummary migrationsResponseType = "summary"
	responseTypeFull    migrationsResponseType = "full"
	modeTypeApply       migrationsModeType     = "apply"
	modeTypeSync        migrationsModeType     = "sync"
	modeTypeDryRun      migrationsModeType     = "dry-run"
)

var modeTypeValidator = func(mode migrationsModeType) bool {
	return mode == "" || mode == modeTypeApply || mode == modeTypeSync || mode == modeTypeDryRun
}

type migrationsPostRequest struct {
	Response migrationsResponseType `json:"response,omitempty"`
	Mode     migrationsModeType     `json:"mode,omitempty"`
}

type tenantsPostRequest struct {
	Name string `json:"name" binding:"required"`
	migrationsPostRequest
}

type migrationsSuccessResponse struct {
	Results           *types.MigrationResults `json:"results,omitempty"`
	AppliedMigrations []types.Migration       `json:"appliedMigrations,omitempty"`
}

type errorResponse struct {
	ErrorMessage string      `json:"error,omitempty"`
	Details      interface{} `json:"details,omitempty"`
}

func getPort(config *config.Config) string {
	if len(strings.TrimSpace(config.Port)) == 0 {
		return defaultPort
	}
	return config.Port
}

func sendNotification(ctx context.Context, config *config.Config, text string) {
	notifier := notifications.NewNotifier(config)
	resp, err := notifier.Notify(text)

	if err != nil {
		common.LogError(ctx, "error=Notifier details=\"%v\"", err)
	} else {
		common.LogInfo(ctx, "success=Notifier response=\"%v\"", resp)
	}
}

func createAndInitLoader(config *config.Config, newLoader func(*config.Config) loader.Loader) loader.Loader {
	loader := newLoader(config)
	return loader
}

func createAndInitConnector(config *config.Config, newConnector func(*config.Config) (db.Connector, error)) (db.Connector, error) {
	connector, err := newConnector(config)
	if err != nil {
		return nil, err
	}
	if err := connector.Init(); err != nil {
		return nil, err
	}
	return connector, nil
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

func requestLoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		common.LogInfo(c.Request.Context(), "clientIP=%v method=%v request=%v", c.ClientIP(), c.Request.Method, c.Request.URL.RequestURI())
		c.Next()
	}
}

func makeHandler(handler func(c *gin.Context, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader), config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler(c, config, newConnector, newLoader)
	}
}

func configHandler(c *gin.Context, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	c.YAML(200, config)
}

func diskMigrationsHandler(c *gin.Context, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	loader := createAndInitLoader(config, newLoader)
	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		common.LogError(c.Request.Context(), "Error getting disk migrations: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, &errorResponse{err.Error(), nil})
	} else {
		common.LogInfo(c.Request.Context(), "Returning disk migrations: %v", len(diskMigrations))
		c.JSON(http.StatusOK, diskMigrations)
	}
}

func tenantsGetHandler(c *gin.Context, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		common.LogError(c.Request.Context(), "Error creating connector: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, &errorResponse{err.Error(), nil})
		return
	}
	defer connector.Dispose()
	tenants, err := connector.GetTenants()
	if err != nil {
		common.LogError(c.Request.Context(), "Error getting tenants: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, &errorResponse{err.Error(), nil})
		return
	}
	common.LogInfo(c.Request.Context(), "Returning tenants: %v", len(tenants))
	c.JSON(http.StatusOK, tenants)
}

func tenantsPostHandler(c *gin.Context, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	var tenant tenantsPostRequest
	err := c.ShouldBindJSON(&tenant)
	if err != nil {
		common.LogError(c.Request.Context(), "Bad request: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{err.Error(), nil})
		return
	}

	loader := createAndInitLoader(config, newLoader)
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		common.LogError(c.Request.Context(), "Error creating connector: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}
	defer connector.Dispose()

	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		common.LogError(c.Request.Context(), "Error getting disk migrations: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}

	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		common.LogError(c.Request.Context(), "Error getting DB migrations: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}

	verified, offendingMigrations := migrations.VerifyCheckSums(diskMigrations, dbMigrations)

	if !verified {
		common.LogError(c.Request.Context(), "error=ChecksumVerificationFailed numberOfOffendingMigrations=%v", len(offendingMigrations))
		c.AbortWithStatusJSON(http.StatusFailedDependency, errorResponse{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		return
	}

	// filter only tenant schemas
	migrationsToApply := migrations.FilterTenantMigrations(c.Request.Context(), diskMigrations)
	common.LogInfo(c.Request.Context(), "Found migrations to apply: %d", len(migrationsToApply))

	results, err := connector.AddTenantAndApplyMigrations(c.Request.Context(), tenant.Name, migrationsToApply)
	if err != nil {
		common.LogError(c.Request.Context(), "Error adding new tenant: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}

	text := fmt.Sprintf("Tenant %q added, migrations applied: %v", tenant.Name, len(migrationsToApply))
	sendNotification(c.Request.Context(), config, text)

	common.LogInfo(c.Request.Context(), text)
	c.JSON(http.StatusOK, migrationsSuccessResponse{results, migrationsToApply})
}

func migrationsGetHandler(c *gin.Context, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		common.LogError(c.Request.Context(), "error=ConnectorInit details=\"%v\"", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}
	defer connector.Dispose()
	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		common.LogError(c.Request.Context(), "error=ConnectorGetDBMigrations details=\"%v\"", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}
	common.LogInfo(c.Request.Context(), "success=ReturningMigrations migrations=%v", len(dbMigrations))
	c.JSON(http.StatusOK, dbMigrations)
}

func migrationsPostHandler(c *gin.Context, config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) {
	var request migrationsPostRequest

	if c.Request.Body != nil {
		err := c.ShouldBindJSON(&request)
		if err != nil {
			common.LogError(c.Request.Context(), "Error reading request: %v", err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
			return
		}
	}

	if modeTypeValidator(request.Mode) == false {
		c.AbortWithStatusJSON(http.StatusBadRequest, errorResponse{fmt.Sprintf("Valid mode parameters are: %v, %v, %v", modeTypeApply, modeTypeSync, modeTypeDryRun), nil})
		return
	}

	if request.Response == "" {
		request.Response = responseTypeSummary
	}
	if request.Mode == "" {
		request.Mode = modeTypeApply
	}

	loader := createAndInitLoader(config, newLoader)
	connector, err := createAndInitConnector(config, newConnector)
	if err != nil {
		common.LogError(c.Request.Context(), "Error creating connector: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}
	defer connector.Dispose()

	diskMigrations, err := loader.GetDiskMigrations()
	if err != nil {
		common.LogError(c.Request.Context(), "Error getting disk migrations: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}

	dbMigrations, err := connector.GetDBMigrations()
	if err != nil {
		common.LogError(c.Request.Context(), "Error getting DB migrations: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}

	verified, offendingMigrations := migrations.VerifyCheckSums(diskMigrations, dbMigrations)
	if !verified {
		common.LogError(c.Request.Context(), "Checksum verification failed for migrations: %v", len(offendingMigrations))
		c.AbortWithStatusJSON(http.StatusFailedDependency, errorResponse{"Checksum verification failed. Please review offending migrations.", offendingMigrations})
		return
	}

	migrationsToApply := migrations.ComputeMigrationsToApply(c.Request.Context(), diskMigrations, dbMigrations)
	common.LogInfo(c.Request.Context(), "Found migrations to apply: %d", len(migrationsToApply))

	results, err := connector.ApplyMigrations(c.Request.Context(), migrationsToApply)
	if err != nil {
		common.LogError(c.Request.Context(), "Error applying migrations: %v", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, errorResponse{err.Error(), nil})
		return
	}

	text := fmt.Sprintf("Applied migrations: %v", len(migrationsToApply))
	sendNotification(c.Request.Context(), config, text)

	common.LogInfo(c.Request.Context(), "Returning applied migrations: %v", len(migrationsToApply))

	c.JSON(http.StatusOK, migrationsSuccessResponse{results, migrationsToApply})
}

func setupRouter(config *config.Config, newConnector func(*config.Config) (db.Connector, error), newLoader func(*config.Config) loader.Loader) *gin.Engine {
	r := gin.New()
	r.HandleMethodNotAllowed = true
	r.Use(gin.Recovery(), requestIDHandler(), requestLoggerHandler())

	v1 := r.Group("/v1")

	v1.GET("/config", makeHandler(configHandler, config, newConnector, newLoader))
	v1.GET("/diskMigrations", makeHandler(diskMigrationsHandler, config, newConnector, newLoader))

	v1.GET("/tenants", makeHandler(tenantsGetHandler, config, newConnector, newLoader))
	v1.POST("/tenants", makeHandler(tenantsPostHandler, config, newConnector, newLoader))

	v1.GET("/migrations", makeHandler(migrationsGetHandler, config, newConnector, newLoader))
	v1.POST("/migrations", makeHandler(migrationsPostHandler, config, newConnector, newLoader))

	return r
}
