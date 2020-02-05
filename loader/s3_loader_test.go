package loader

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/lukaszbudnik/migrator/config"
	"github.com/stretchr/testify/assert"
)

type mockS3Client struct {
	s3iface.S3API
}

func (m *mockS3Client) ListObjectsV2Pages(input *s3.ListObjectsV2Input, callback func(*s3.ListObjectsV2Output, bool) bool) error {

	var contents []*s3.Object

	switch *input.Prefix {
	case "migrations/config":
		file1 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "201602160001.sql"))}
		file2 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "201602160002.sql"))}
		contents = []*s3.Object{file1, file2}
	case "migrations/ref":
		file1 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "202001100003.sql"))}
		file2 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "202001100005.sql"))}
		contents = []*s3.Object{file1, file2}
	case "migrations/tenants":
		file1 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "201602160002.sql"))}
		file2 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "202001100004.sql"))}
		file3 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "202001100007.sql"))}
		contents = []*s3.Object{file1, file2, file3}
	case "migrations/config-scripts":
		file1 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "recreate-triggers.sql"))}
		file2 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "cleanup.sql"))}
		contents = []*s3.Object{file1, file2}
	case "migrations/tenants-scripts":
		file1 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "recreate-triggers.sql"))}
		file2 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "cleanup.sql"))}
		file3 := &s3.Object{Key: aws.String(fmt.Sprintf("%v/%v", *input.Prefix, "run-reports.sql"))}
		contents = []*s3.Object{file1, file2, file3}
	}

	callback(&s3.ListObjectsV2Output{
		Contents: contents,
		KeyCount: aws.Int64(int64(len(contents))),
	}, true)
	return nil
}

func (m *mockS3Client) GetObject(input *s3.GetObjectInput) (output *s3.GetObjectOutput, err error) {
	return &s3.GetObjectOutput{Body: ioutil.NopCloser(bytes.NewReader([]byte(*input.Key)))}, nil
}

func TestS3GetSourceMigrations(t *testing.T) {
	mock := &mockS3Client{}

	config := &config.Config{
		BaseLocation:     "s3://lukasz-budnik-migrator-us-east-1",
		SingleMigrations: []string{"migrations/config", "migrations/ref"},
		TenantMigrations: []string{"migrations/tenants"},
		SingleScripts:    []string{"migrations/config-scripts"},
		TenantScripts:    []string{"migrations/tenants-scripts"},
	}

	loader := &s3Loader{baseLoader{context.TODO(), config}}
	migrations := loader.doGetSourceMigrations(mock)

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
