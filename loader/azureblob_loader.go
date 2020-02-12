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

	"github.com/lukaszbudnik/migrator/types"
)

// azureBlobLoader is struct used for implementing Loader interface for loading migrations from Azure Blob
type azureBlobLoader struct {
	baseLoader
}

// GetSourceMigrations returns all migrations from Azure Blob location
func (abl *azureBlobLoader) GetSourceMigrations() []types.Migration {
	accountName, accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT"), os.Getenv("AZURE_STORAGE_ACCESS_KEY")

	if len(accountName) == 0 || len(accountKey) == 0 {
		panic("Either the AZURE_STORAGE_ACCOUNT or AZURE_STORAGE_ACCESS_KEY environment variable is not set")
	}

	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)

	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})

	u, err := url.Parse(abl.config.BaseLocation)
	if err != nil {
		panic(err.Error())
	}

	// migrator expects that container as a part of the service url
	// for example: https://lukaszbudniktest.blob.core.windows.net/mycontainer
	// Azure SDK will correctly parse the account and the container
	containerURL := azblob.NewContainerURL(*u, p)

	return abl.doGetSourceMigrations(containerURL)
}

func (abl *azureBlobLoader) doGetSourceMigrations(containerURL azblob.ContainerURL) []types.Migration {
	migrations := []types.Migration{}

	singleMigrationsObjects := abl.getObjectList(containerURL, abl.config.SingleMigrations)
	tenantMigrationsObjects := abl.getObjectList(containerURL, abl.config.TenantMigrations)
	singleScriptsObjects := abl.getObjectList(containerURL, abl.config.SingleScripts)
	tenantScriptsObjects := abl.getObjectList(containerURL, abl.config.TenantScripts)

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

func (abl *azureBlobLoader) getObjectList(containerURL azblob.ContainerURL, prefixes []string) []string {
	objects := []string{}

	for _, prefix := range prefixes {

		for marker := (azblob.Marker{}); marker.NotDone(); { // The parens around Marker{} are required to avoid compiler error.

			listBlob, err := containerURL.ListBlobsFlatSegment(abl.ctx, marker, azblob.ListBlobsSegmentOptions{Prefix: prefix + "/"})
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

		get, err := blobURL.Download(abl.ctx, 0, 0, azblob.BlobAccessConditions{}, false)
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
