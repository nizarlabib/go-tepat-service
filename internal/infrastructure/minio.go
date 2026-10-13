package infrastructure

import (
	"log"
	"log/slog"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/spf13/viper"
)

type Minio interface {
	GetClient() *minio.Client
	GetBucketName() string
}

type minioClient struct {
	client *minio.Client
	bucket string
}

func NewMinio() Minio {
	viper.SetDefault("MINIO_REGION", "ap-southeast-1")
	viper.SetDefault("MINIO_USE_SSL", false)

	endpoint := viper.GetString("MINIO_ENDPOINT")
	accessKey := viper.GetString("MINIO_ACCESS_KEY")
	secretKey := viper.GetString("MINIO_SECRET_KEY")
	region := viper.GetString("MINIO_REGION")
	bucket := viper.GetString("MINIO_BUCKET")
	useSSL := viper.GetBool("MINIO_USE_SSL")

	client, err := minio.New(endpoint, &minio.Options{
		Secure: useSSL,
		Region: region,
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
	})
	if err != nil {
		log.Fatal("Failed to initiate Minio:", err)
	}
	slog.Info("Connected to Minio Success . . .")
	return minioClient{
		client: client,
		bucket: bucket,
	}
}

func (m minioClient) GetClient() *minio.Client {
	return m.client
}

func (m minioClient) GetBucketName() string {
	return m.bucket
}
