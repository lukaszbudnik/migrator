package loader

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/azure/auth"

	"github.com/lukaszbudnik/migrator/common"
	"github.com/lukaszbudnik/migrator/types"
)

// azureBlobLoader is struct used for implementing Loader interface for loading migrations from Azure Blob
type azureBlobLoader struct {
	baseLoader
}

// GetSourceMigrations returns all migrations from Azure Blob location
func (abl *azureBlobLoader) GetSourceMigrations() []types.Migration {

	credential, err := abl.getAzureStorageCredentials()
	if err != nil {
		panic(err.Error())
	}

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	// migrator expects that container as a part of the service url
	// the URL can contain optional prefixes like prod/artefacts
	// for example:
	// https://lukaszbudniktest.blob.core.windows.net/mycontainer/
	// https://lukaszbudniktest.blob.core.windows.net/mycontainer/prod/artefacts/

	// check if optional prefixes are provided
	baseLocation := strings.TrimRight(abl.config.BaseLocation, "/")
	indx := common.FindNthIndex(baseLocation, '/', 4)

	optionalPrefixes := ""
	if indx > -1 {
		optionalPrefixes = baseLocation[indx+1:]
		baseLocation = baseLocation[:indx]
	}

	u, err := url.Parse(baseLocation)
	if err != nil {
		panic(err.Error())
	}

	containerURL := azblob.NewContainerURL(*u, p)

	return abl.doGetSourceMigrations(containerURL, optionalPrefixes)
}

func (abl *azureBlobLoader) doGetSourceMigrations(containerURL azblob.ContainerURL, optionalPrefixes string) []types.Migration {
	migrations := []types.Migration{}

	singleMigrationsObjects := abl.getObjectList(containerURL, optionalPrefixes, abl.config.SingleMigrations)
	tenantMigrationsObjects := abl.getObjectList(containerURL, optionalPrefixes, abl.config.TenantMigrations)
	singleScriptsObjects := abl.getObjectList(containerURL, optionalPrefixes, abl.config.SingleScripts)
	tenantScriptsObjects := abl.getObjectList(containerURL, optionalPrefixes, abl.config.TenantScripts)

	migrationsMap := make(map[string][]types.Migration)
	abl.getObjects(containerURL, migrationsMap, singleMigrationsObjects, types.MigrationTypeSingleMigration)
	abl.getObjects(containerURL, migrationsMap, tenantMigrationsObjects, types.MigrationTypeTenantMigration)
	abl.sortMigrations(migrationsMap, &migrations)

	migrationsMap = make(map[string][]types.Migration)
	abl.getObjects(containerURL, migrationsMap, singleScriptsObjects, types.MigrationTypeSingleScript)
	abl.sortMigrations(migrationsMap, &migrations)

	migrationsMap = make(map[string][]types.Migration)
	abl.getObjects(containerURL, migrationsMap, tenantScriptsObjects, types.MigrationTypeTenantScript)
	abl.sortMigrations(migrationsMap, &migrations)

	return migrations
}

func (abl *azureBlobLoader) getObjectList(containerURL azblob.ContainerURL, optionalPrefixes string, prefixes []string) []string {
	objects := []string{}

	for _, prefix := range prefixes {

		for marker := (azblob.Marker{}); marker.NotDone(); {

			var fullPrefix string
			if optionalPrefixes != "" {
				fullPrefix = optionalPrefixes + "/" + prefix + "/"
			} else {
				fullPrefix = prefix + "/"
			}

			listBlob, err := containerURL.ListBlobsFlatSegment(abl.ctx, marker, azblob.ListBlobsSegmentOptions{Prefix: fullPrefix})
			if err != nil {
				panic(err.Error())
			}
			marker = listBlob.NextMarker

			for _, blobInfo := range listBlob.Segment.BlobItems {
				objects = append(objects, blobInfo.Name)
			}
		}

	}

	return objects
}

func (abl *azureBlobLoader) getObjects(containerURL azblob.ContainerURL, migrationsMap map[string][]types.Migration, objects []string, migrationType types.MigrationType) {
	for _, o := range objects {
		blobURL := containerURL.NewBlobURL(o)

		get, err := blobURL.Download(abl.ctx, 0, 0, azblob.BlobAccessConditions{}, false, azblob.ClientProvidedKeyOptions{})
		if err != nil {
			panic(err.Error())
		}

		downloadedData := &bytes.Buffer{}
		reader := get.Body(azblob.RetryReaderOptions{})
		downloadedData.ReadFrom(reader)
		reader.Close()

		contents := downloadedData.String()

		hasher := sha256.New()
		hasher.Write([]byte(contents))
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

func (abl *azureBlobLoader) getAzureStorageCredentials() (azblob.Credential, error) {
	// try shared key credentials first
	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")

	if len(accountName) > 0 && len(accountKey) > 0 {
		return azblob.NewSharedKeyCredential(accountName, accountKey)
	}

	// then try MSI and token credentials
	msiConfig := auth.NewMSIConfig()
	msiConfig.Resource = "https://storage.azure.com"

	azureServicePrincipalToken, err := msiConfig.ServicePrincipalToken()
	if err != nil {
		return nil, err
	}

	err = azureServicePrincipalToken.Refresh()
	if err != nil {
		return nil, err
	}
	token := azureServicePrincipalToken.Token()

	credential := azblob.NewTokenCredential(token.AccessToken, nil)
	return credential, nil
}
