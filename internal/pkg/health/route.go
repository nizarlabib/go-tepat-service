package health

import (
	"log/slog"
	"net/http"
	"strings"

	"go-memolab-service/internal/infrastructure"
	// "github.com/minio/minio-go/v7"
	// "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

func NewHealthRoute() *healthRoute {
	serviceName := viper.GetString("APP_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "unknown-service"
	}
	serviceVersion := viper.GetString("APP_VERSION")
	if serviceVersion == "" {
		serviceVersion = "dev"
	}

	h, _ := New(
		WithComponent(Component{
			Name:    serviceName,
			Version: serviceVersion,
		}),
		WithSystemInfo(),
		WithOutputChecks(),
	)

	return &healthRoute{
		handler: h,
	}
}

type healthRoute struct {
	handler *Health
}

func (r *healthRoute) Path() string {
	return "/health"
}

func (r *healthRoute) Register(resource ...any) error {
	for _, res := range resource {
		if err := r.register(res); err != nil {
			return err
		}
	}
	return nil
}

func (r *healthRoute) register(resource any) error {
	var config Config

	switch res := resource.(type) {
	case Config:
		config = res
	case *gorm.DB:
		config = PostgreSQLHealthConfig(res)
	case infrastructure.Postgres:
		config = PostgreSQLHealthConfig(res.GetClientGorm())
	case *redis.Client:
		config = RedisHealthConfig(res)
	case infrastructure.Redis:
		config = RedisHealthConfig(res.GetClient())
	// case *minio.Client:
	// 	config = MinioHealthConfig(res, "")
	// case infrastructure.Minio:
	// 	config = MinioHealthConfig(res.GetClient(), res.GetBucketName())
	// case *amqp091.Connection:
	// 	config = RabbitMQHealthConfig(res)
	// case rabbitmq.RabbitMq:
	// 	config = RabbitMQHealthConfig(res.GetConnection())
	default:
		return nil
	}

	if err := r.handler.Register(config); err != nil {
		return err
	}

	slog.Debug("registered health check component", "name", config.Name)

	return nil
}

func (r *healthRoute) Middleware(endpoint string) func(http.Handler) http.Handler {
	f := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, re *http.Request) {
			if re.Method == "GET" && strings.EqualFold(re.URL.Path, endpoint) {
				r.handler.HandlerFunc(w, re)
				return
			}
			h.ServeHTTP(w, re)
		}
		return http.HandlerFunc(fn)
	}
	return f
}

func (r *healthRoute) Handler() http.Handler {
	return r.handler.Handler()
}

func (r *healthRoute) HandlerFunc(w http.ResponseWriter, re *http.Request) {
	r.handler.HandlerFunc(w, re)
}
