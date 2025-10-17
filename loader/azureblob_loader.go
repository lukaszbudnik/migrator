package loader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"

	"github.com/lukaszbudnik/migrator/types"
)

// AzureBlobClient interface for Azure Blob operations
type AzureBlobClient interface {
	NewListBlobsFlatPager(containerName string, options *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse]
	DownloadStream(ctx context.Context, containerName, blobName string, options *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error)
}

// AzureBlobClientFactory creates Azure Blob clients
type AzureBlobClientFactory interface {
	NewClient(ctx context.Context, serviceURL, containerName string) (AzureBlobClient, error)
}

// azureBlobLoader is struct used for implementing Loader interface for loading migrations from Azure Blob
type azureBlobLoader struct {
	baseLoader
	clientFactory AzureBlobClientFactory
}

// defaultAzureBlobClientFactory implements AzureBlobClientFactory
type defaultAzureBlobClientFactory struct{}

func (f *defaultAzureBlobClientFactory) NewClient(ctx context.Context, serviceURL, containerName string) (AzureBlobClient, error) {
	// then try managed identity
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, err
	}
	client, err := azblob.NewClient(serviceURL, credential, nil)
	if err != nil {
		return nil, err
	}

	return &azureBlobClientWrapper{client: client, containerName: containerName}, nil
}

// azureBlobClientWrapper wraps the Azure Blob client to implement our interface
type azureBlobClientWrapper struct {
	client        *azblob.Client
	containerName string
}

func (w *azureBlobClientWrapper) NewListBlobsFlatPager(containerName string, options *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse] {
	return w.client.NewListBlobsFlatPager(containerName, options)
}

func (w *azureBlobClientWrapper) DownloadStream(ctx context.Context, containerName, blobName string, options *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error) {
	return w.client.DownloadStream(ctx, containerName, blobName, options)
}

func (abl *azureBlobLoader) getClientFactory() AzureBlobClientFactory {
	if abl.clientFactory != nil {
		return abl.clientFactory
	}
	return &defaultAzureBlobClientFactory{}
}

// GetSourceMigrations returns all migrations from Azure Blob location
func (abl *azureBlobLoader) GetSourceMigrations() []types.Migration {
	// migrator expects that container as a part of the service url
	// the URL can contain optional prefixes like prod/artefacts
	// for example:
	// https://lukaszbudniktest.blob.core.windows.net/mycontainer/
	// https://lukaszbudniktest.blob.core.windows.net/mycontainer/prod/artefacts/

	// Parse URL to extract service URL and container name

	serviceURL, containerName, optionalPrefixes := abl.parseBaseLocation()

	client, err := abl.getClientFactory().NewClient(abl.ctx, serviceURL, containerName)
	if err != nil {
		panic(err.Error())
	}

	return abl.doGetSourceMigrations(client, containerName, optionalPrefixes)
}

func (abl *azureBlobLoader) doGetSourceMigrations(client AzureBlobClient, containerName, optionalPrefixes string) []types.Migration {
	migrations := []types.Migration{}

	singleMigrationsObjects := abl.getObjectList(client, containerName, optionalPrefixes, abl.config.SingleMigrations)
	tenantMigrationsObjects := abl.getObjectList(client, containerName, optionalPrefixes, abl.config.TenantMigrations)
	singleScriptsObjects := abl.getObjectList(client, containerName, optionalPrefixes, abl.config.SingleScripts)
	tenantScriptsObjects := abl.getObjectList(client, containerName, optionalPrefixes, abl.config.TenantScripts)

	migrationsMap := make(map[string][]types.Migration)
	abl.getObjects(client, containerName, migrationsMap, singleMigrationsObjects, types.MigrationTypeSingleMigration)
	abl.getObjects(client, containerName, migrationsMap, tenantMigrationsObjects, types.MigrationTypeTenantMigration)
	abl.sortMigrations(migrationsMap, &migrations)

	migrationsMap = make(map[string][]types.Migration)
	abl.getObjects(client, containerName, migrationsMap, singleScriptsObjects, types.MigrationTypeSingleScript)
	abl.sortMigrations(migrationsMap, &migrations)

	migrationsMap = make(map[string][]types.Migration)
	abl.getObjects(client, containerName, migrationsMap, tenantScriptsObjects, types.MigrationTypeTenantScript)
	abl.sortMigrations(migrationsMap, &migrations)

	return migrations
}

func (abl *azureBlobLoader) getObjectList(client AzureBlobClient, containerName, optionalPrefixes string, prefixes []string) []string {
	objects := []string{}

	for _, prefix := range prefixes {
		var fullPrefix string
		if optionalPrefixes != "" {
			fullPrefix = optionalPrefixes + "/" + prefix + "/"
		} else {
			fullPrefix = prefix + "/"
		}

		pager := client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
			Prefix: &fullPrefix,
		})

		for pager.More() {
			page, err := pager.NextPage(abl.ctx)
			if err != nil {
				panic(err.Error())
			}

			for _, blob := range page.Segment.BlobItems {
				if blob.Name != nil {
					objects = append(objects, *blob.Name)
				}
			}
		}
	}

	return objects
}

func (abl *azureBlobLoader) getObjects(client AzureBlobClient, containerName string, migrationsMap map[string][]types.Migration, objects []string, migrationType types.MigrationType) {
	for _, o := range objects {
		response, err := client.DownloadStream(abl.ctx, containerName, o, nil)
		if err != nil {
			panic(err.Error())
		}

		contents, err := io.ReadAll(response.Body)
		if err != nil {
			panic(err.Error())
		}
		response.Body.Close()

		hasher := sha256.New()
		hasher.Write(contents)
		file := fmt.Sprintf("%s/%s", abl.config.BaseLocation, o)
		from := strings.LastIndex(file, "/")
		sourceDir := file[0:from]
		name := file[from+1:]
		m := types.Migration{Name: name, SourceDir: sourceDir, File: file, MigrationType: migrationType, Contents: string(contents), CheckSum: hex.EncodeToString(hasher.Sum(nil))}

		e, ok := migrationsMap[m.Name]
		if ok {
			e = append(e, m)
		} else {
			e = []types.Migration{m}
		}
		migrationsMap[m.Name] = e
	}
}

func (abl *azureBlobLoader) HealthCheck() error {
	serviceURL, containerName, prefix := abl.parseBaseLocation()

	client, err := abl.getClientFactory().NewClient(abl.ctx, serviceURL, containerName)
	if err != nil {
		return err
	}

	pager := client.NewListBlobsFlatPager(containerName, &azblob.ListBlobsFlatOptions{
		Prefix: &prefix,
	})

	if pager.More() {
		_, err = pager.NextPage(abl.ctx)
		return err
	}

	return nil
}

func (abl *azureBlobLoader) parseBaseLocation() (string, string, string) {
	baseLocation := strings.TrimSpace(abl.config.BaseLocation)
	u, err := url.Parse(baseLocation)
	if err != nil {
		panic(err)
	}

	serviceURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)

	var containerName string
	var optionalPrefixes string

	pathComponents := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathComponents) > 0 {
		containerName = pathComponents[0]
	}

	if len(pathComponents) > 1 {
		optionalPrefixes = strings.Join(pathComponents[1:], "/")
	}

	return serviceURL, containerName, optionalPrefixes
}
