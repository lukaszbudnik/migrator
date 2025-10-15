package loader

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/lukaszbudnik/migrator/types"
)

// S3APIClient interface for S3 operations
type S3APIClient interface {
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
	ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// S3ListObjectsV2Paginator interface for pagination
type S3ListObjectsV2Paginator interface {
	HasMorePages() bool
	NextPage(context.Context, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// S3ClientFactory creates S3 clients
type S3ClientFactory interface {
	NewClient(ctx context.Context) S3APIClient
}

// S3PaginatorFactory creates paginators
type S3PaginatorFactory interface {
	NewListObjectsV2Paginator(client S3APIClient, input *s3.ListObjectsV2Input) S3ListObjectsV2Paginator
}

// s3Loader is struct used for implementing Loader interface for loading migrations from AWS S3
type s3Loader struct {
	baseLoader
	clientFactory    S3ClientFactory
	paginatorFactory S3PaginatorFactory
}

// defaultS3ClientFactory implements S3ClientFactory
type defaultS3ClientFactory struct{}

func (f *defaultS3ClientFactory) NewClient(ctx context.Context) S3APIClient {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err.Error())
	}
	return s3.NewFromConfig(cfg)
}

// defaultS3PaginatorFactory implements S3PaginatorFactory
type defaultS3PaginatorFactory struct{}

func (f *defaultS3PaginatorFactory) NewListObjectsV2Paginator(client S3APIClient, input *s3.ListObjectsV2Input) S3ListObjectsV2Paginator {
	return s3.NewListObjectsV2Paginator(client, input)
}

func (s3l *s3Loader) getPaginatorFactory() S3PaginatorFactory {
	if s3l.paginatorFactory != nil {
		return s3l.paginatorFactory
	}
	return &defaultS3PaginatorFactory{}
}

func (s3l *s3Loader) getClientFactory() S3ClientFactory {
	if s3l.clientFactory != nil {
		return s3l.clientFactory
	}
	return &defaultS3ClientFactory{}
}

// GetSourceMigrations returns all migrations from AWS S3 location
func (s3l *s3Loader) GetSourceMigrations() []types.Migration {
	client := s3l.getClientFactory().NewClient(s3l.ctx)
	return s3l.doGetSourceMigrations(client)
}

func (s3l *s3Loader) HealthCheck() error {
	client := s3l.getClientFactory().NewClient(s3l.ctx)
	return s3l.doHealthCheck(client)
}

func (s3l *s3Loader) doHealthCheck(client S3APIClient) error {
	bucketWithPrefixes := strings.Split(strings.Replace(strings.TrimRight(s3l.config.BaseLocation, "/"), "s3://", "", 1), "/")

	bucket := bucketWithPrefixes[0]
	prefix := "/"
	if len(bucketWithPrefixes) > 1 {
		prefix = strings.Join(bucketWithPrefixes[1:], "/")
	}

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1),
	}

	_, err := client.ListObjectsV2(s3l.ctx, input)

	return err
}

func (s3l *s3Loader) doGetSourceMigrations(client S3APIClient) []types.Migration {
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

func (s3l *s3Loader) getObjectList(client S3APIClient, bucket, optionalPrefixes string, prefixes []string) []*string {
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
			MaxKeys: aws.Int32(100),
		}

		paginator := s3l.getPaginatorFactory().NewListObjectsV2Paginator(client, input)
		for paginator.HasMorePages() {
			page, err := paginator.NextPage(s3l.ctx)
			if err != nil {
				panic(err.Error())
			}
			for _, obj := range page.Contents {
				objects = append(objects, obj.Key)
			}
		}
	}

	return objects
}

func (s3l *s3Loader) getObjects(client S3APIClient, bucket string, migrationsMap map[string][]types.Migration, objects []*string, migrationType types.MigrationType) {

	for _, o := range objects {
		input := &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(*o),
		}
		object, err := client.GetObject(s3l.ctx, input)
		if err != nil {
			panic(err.Error())
		}

		contents, err := io.ReadAll(object.Body)
		if err != nil {
			panic(err.Error())
		}

		hasher := sha256.New()
		hasher.Write(contents)
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
