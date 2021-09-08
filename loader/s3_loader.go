package loader

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/lukaszbudnik/migrator/types"
)

// s3Loader is struct used for implementing Loader interface for loading migrations from AWS S3
type s3Loader struct {
	baseLoader
}

// GetSourceMigrations returns all migrations from AWS S3 location
func (s3l *s3Loader) GetSourceMigrations() []types.Migration {
	sess, err := session.NewSession()
	if err != nil {
		panic(err.Error())
	}
	client := s3.New(sess)
	return s3l.doGetSourceMigrations(client)
}

func (s3l *s3Loader) HealthCheck() error {
	sess, err := session.NewSession()
	if err != nil {
		return err
	}
	client := s3.New(sess)
	return s3l.doHealthCheck(client)
}

func (s3l *s3Loader) doHealthCheck(client s3iface.S3API) error {
	bucketWithPrefixes := strings.Split(strings.Replace(strings.TrimRight(s3l.config.BaseLocation, "/"), "s3://", "", 1), "/")

	bucket := bucketWithPrefixes[0]
	prefix := "/"
	if len(bucketWithPrefixes) > 1 {
		prefix = strings.Join(bucketWithPrefixes[1:], "/")
	}

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int64(1),
	}

	_, err := client.ListObjectsV2(input)

	return err
}

func (s3l *s3Loader) doGetSourceMigrations(client s3iface.S3API) []types.Migration {
	migrations := []types.Migration{}

	bucketWithPrefixes := strings.Split(strings.Replace(strings.TrimRight(s3l.config.BaseLocation, "/"), "s3://", "", 1), "/")

	bucket := bucketWithPrefixes[0]
	optionalPrefixes := ""
	if len(bucketWithPrefixes) > 1 {
		optionalPrefixes = strings.Join(bucketWithPrefixes[1:], "/")
	}

	singleMigrationsObjects := s3l.getObjectList(client, bucket, optionalPrefixes, s3l.config.SingleMigrations)
	tenantMigrationsObjects := s3l.getObjectList(client, bucket, optionalPrefixes, s3l.config.TenantMigrations)
	singleScriptsObjects := s3l.getObjectList(client, bucket, optionalPrefixes, s3l.config.SingleScripts)
	tenantScriptsObjects := s3l.getObjectList(client, bucket, optionalPrefixes, s3l.config.TenantScripts)

	migrationsMap := make(map[string][]types.Migration)
	s3l.getObjects(client, bucket, migrationsMap, singleMigrationsObjects, types.MigrationTypeSingleMigration)
	s3l.getObjects(client, bucket, migrationsMap, tenantMigrationsObjects, types.MigrationTypeTenantMigration)
	s3l.sortMigrations(migrationsMap, &migrations)

	migrationsMap = make(map[string][]types.Migration)
	s3l.getObjects(client, bucket, migrationsMap, singleScriptsObjects, types.MigrationTypeSingleScript)
	s3l.sortMigrations(migrationsMap, &migrations)

	migrationsMap = make(map[string][]types.Migration)
	s3l.getObjects(client, bucket, migrationsMap, tenantScriptsObjects, types.MigrationTypeTenantScript)
	s3l.sortMigrations(migrationsMap, &migrations)

	return migrations
}

func (s3l *s3Loader) getObjectList(client s3iface.S3API, bucket, optionalPrefixes string, prefixes []string) []*string {
	objects := []*string{}

	for _, prefix := range prefixes {

		var fullPrefix string
		if optionalPrefixes != "" {
			fullPrefix = optionalPrefixes + "/" + prefix
		} else {
			fullPrefix = prefix
		}

		input := &s3.ListObjectsV2Input{
			Bucket:  aws.String(bucket),
			Prefix:  aws.String(fullPrefix),
			MaxKeys: aws.Int64(1000),
		}

		err := client.ListObjectsV2Pages(input,
			func(page *s3.ListObjectsV2Output, lastPage bool) bool {

				for _, o := range page.Contents {
					objects = append(objects, o.Key)
				}

				return !lastPage
			})

		if err != nil {
			panic(err.Error())
		}
	}

	return objects
}

func (s3l *s3Loader) getObjects(client s3iface.S3API, bucket string, migrationsMap map[string][]types.Migration, objects []*string, migrationType types.MigrationType) {
	objectInput := &s3.GetObjectInput{Bucket: aws.String(bucket)}
	for _, o := range objects {
		objectInput.Key = o
		objectOutput, err := client.GetObject(objectInput)
		if err != nil {
			panic(err.Error())
		}
		buf := new(bytes.Buffer)
		buf.ReadFrom(objectOutput.Body)
		contents := buf.String()

		hasher := sha256.New()
		hasher.Write([]byte(contents))
		file := fmt.Sprintf("%s/%s", s3l.config.BaseLocation, *o)
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
