package loader

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

type mockS3Client struct {
	S3APIClient
}

func (m *mockS3Client) ListObjectsV2(context.Context, *s3.ListObjectsV2Input, ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	return &s3.ListObjectsV2Output{}, nil
}

func (m *mockS3Client) GetObject(ctx context.Context, input *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte(*input.Key)))}, nil
}

type mockS3ClientFactory struct {
	client S3APIClient
}

func (f *mockS3ClientFactory) NewClient(ctx context.Context) S3APIClient {
	return f.client
}

type mockS3PaginatorFactory struct{}

func (f *mockS3PaginatorFactory) NewListObjectsV2Paginator(client S3APIClient, input *s3.ListObjectsV2Input) S3ListObjectsV2Paginator {
	return &mockS3Paginator{prefix: *input.Prefix, called: false}
}

type mockS3Paginator struct {
	prefix string
	called bool
}

func (m *mockS3Paginator) HasMorePages() bool {
	return !m.called
}

func (m *mockS3Paginator) NextPage(ctx context.Context, optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error) {
	m.called = true

	var contents []types.Object
	switch m.prefix {
	case "migrations/config":
		contents = []types.Object{
			{Key: aws.String("migrations/config/201602160001.sql")},
			{Key: aws.String("migrations/config/201602160002.sql")},
		}
	case "migrations/ref":
		contents = []types.Object{
			{Key: aws.String("migrations/ref/202001100003.sql")},
			{Key: aws.String("migrations/ref/202001100005.sql")},
		}
	case "migrations/tenants":
		contents = []types.Object{
			{Key: aws.String("migrations/tenants/201602160002.sql")},
			{Key: aws.String("migrations/tenants/202001100004.sql")},
			{Key: aws.String("migrations/tenants/202001100007.sql")},
		}
	case "migrations/config-scripts":
		contents = []types.Object{
			{Key: aws.String("migrations/config-scripts/recreate-triggers.sql")},
			{Key: aws.String("migrations/config-scripts/cleanup.sql")},
		}
	case "migrations/tenants-scripts":
		contents = []types.Object{
			{Key: aws.String("migrations/tenants-scripts/recreate-triggers.sql")},
			{Key: aws.String("migrations/tenants-scripts/cleanup.sql")},
			{Key: aws.String("migrations/tenants-scripts/run-reports.sql")},
		}
	case "application-x/prod/migrations/config":
		contents = []types.Object{
			{Key: aws.String("application-x/prod/migrations/config/201602160001.sql")},
			{Key: aws.String("application-x/prod/migrations/config/201602160002.sql")},
		}
	case "application-x/prod/migrations/ref":
		contents = []types.Object{
			{Key: aws.String("application-x/prod/migrations/ref/202001100003.sql")},
			{Key: aws.String("application-x/prod/migrations/ref/202001100005.sql")},
		}
	case "application-x/prod/migrations/tenants":
		contents = []types.Object{
			{Key: aws.String("application-x/prod/migrations/tenants/201602160002.sql")},
			{Key: aws.String("application-x/prod/migrations/tenants/202001100004.sql")},
			{Key: aws.String("application-x/prod/migrations/tenants/202001100007.sql")},
		}
	case "application-x/prod/migrations/config-scripts":
		contents = []types.Object{
			{Key: aws.String("application-x/prod/migrations/config-scripts/recreate-triggers.sql")},
			{Key: aws.String("application-x/prod/migrations/config-scripts/cleanup.sql")},
		}
	case "application-x/prod/migrations/tenants-scripts":
		contents = []types.Object{
			{Key: aws.String("application-x/prod/migrations/tenants-scripts/recreate-triggers.sql")},
			{Key: aws.String("application-x/prod/migrations/tenants-scripts/cleanup.sql")},
			{Key: aws.String("application-x/prod/migrations/tenants-scripts/run-reports.sql")},
		}
	}

	return &s3.ListObjectsV2Output{Contents: contents}, nil
}

func TestS3GetSourceMigrations(t *testing.T) {
	mock := &mockS3Client{}

	config := &config.Config{
		BaseLocation:     "s3://your-bucket-migrator",
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &s3Loader{
		baseLoader:       baseLoader{context.TODO(), config},
		clientFactory:    &mockS3ClientFactory{client: mock},
		paginatorFactory: &mockS3PaginatorFactory{},
	}

	migrations := loader.doGetSourceMigrations(loader.getClientFactory().NewClient(context.TODO()))

	assert.Len(t, migrations, 12)

	assert.Contains(t, migrations[0].File, "migrations/config/201602160001.sql")
	assert.Contains(t, migrations[1].File, "migrations/config/201602160002.sql")
	assert.Contains(t, migrations[2].File, "migrations/tenants/201602160002.sql")
	assert.Contains(t, migrations[3].File, "migrations/ref/202001100003.sql")
	assert.Contains(t, migrations[4].File, "migrations/tenants/202001100004.sql")
	assert.Contains(t, migrations[5].File, "migrations/ref/202001100005.sql")
	assert.Contains(t, migrations[6].File, "migrations/tenants/202001100007.sql")
	assert.Contains(t, migrations[7].File, "migrations/config-scripts/cleanup.sql")
	assert.Contains(t, migrations[8].File, "migrations/config-scripts/recreate-triggers.sql")
	assert.Contains(t, migrations[9].File, "migrations/tenants-scripts/cleanup.sql")
	assert.Contains(t, migrations[10].File, "migrations/tenants-scripts/recreate-triggers.sql")
	assert.Contains(t, migrations[11].File, "migrations/tenants-scripts/run-reports.sql")

}

func TestS3GetSourceMigrationsBucketWithPrefix(t *testing.T) {
	mock := &mockS3Client{}

	config := &config.Config{
		BaseLocation:     "s3://your-bucket-migrator/application-x/prod/",
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &s3Loader{
		baseLoader:       baseLoader{context.TODO(), config},
		clientFactory:    &mockS3ClientFactory{client: mock},
		paginatorFactory: &mockS3PaginatorFactory{},
	}
	migrations := loader.doGetSourceMigrations(loader.getClientFactory().NewClient(context.TODO()))

	assert.Len(t, migrations, 12)

	assert.Contains(t, migrations[0].File, "application-x/prod/migrations/config/201602160001.sql")
	assert.Contains(t, migrations[1].File, "application-x/prod/migrations/config/201602160002.sql")
	assert.Contains(t, migrations[2].File, "application-x/prod/migrations/tenants/201602160002.sql")
	assert.Contains(t, migrations[3].File, "application-x/prod/migrations/ref/202001100003.sql")
	assert.Contains(t, migrations[4].File, "application-x/prod/migrations/tenants/202001100004.sql")
	assert.Contains(t, migrations[5].File, "application-x/prod/migrations/ref/202001100005.sql")
	assert.Contains(t, migrations[6].File, "application-x/prod/migrations/tenants/202001100007.sql")
	assert.Contains(t, migrations[7].File, "application-x/prod/migrations/config-scripts/cleanup.sql")
	assert.Contains(t, migrations[8].File, "application-x/prod/migrations/config-scripts/recreate-triggers.sql")
	assert.Contains(t, migrations[9].File, "application-x/prod/migrations/tenants-scripts/cleanup.sql")
	assert.Contains(t, migrations[10].File, "application-x/prod/migrations/tenants-scripts/recreate-triggers.sql")
	assert.Contains(t, migrations[11].File, "application-x/prod/migrations/tenants-scripts/run-reports.sql")
}

func TestS3HealthCheck(t *testing.T) {
	mock := &mockS3Client{}

	config := &config.Config{
		BaseLocation:     "s3://your-bucket-migrator",
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &s3Loader{
		baseLoader:       baseLoader{context.TODO(), config},
		clientFactory:    &mockS3ClientFactory{client: mock},
		paginatorFactory: &mockS3PaginatorFactory{},
	}
	err := loader.doHealthCheck(loader.getClientFactory().NewClient(context.TODO()))

	assert.Nil(t, err)
}
