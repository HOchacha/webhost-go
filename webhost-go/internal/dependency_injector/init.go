package dependency_injector

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"time"
	"webhost-go/webhost-go/internal/controller"
	"webhost-go/webhost-go/internal/db_driver"
	"webhost-go/webhost-go/internal/middleware"
	"webhost-go/webhost-go/internal/services/hosting_service"
	"webhost-go/webhost-go/internal/services/user_service"
	"webhost-go/webhost-go/internal/services/user_service/authn/token"
	"webhost-go/webhost-go/internal/services/user_service/authn/utils"
	"webhost-go/webhost-go/pkg/aws"
)

type AppInitializer struct {
	DB        DBConfig
	JWTSecret string
	TokenTTL  time.Duration
}

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     int
	Name     string
}

func (ai *AppInitializer) InitApp() (*HandlerRegistry, error) {
	db, err := sql.Open("mysql", ai.DB.DSN())
	if err != nil {
		panic(err)
	}

	userRepo := db_driver.NewUserRepository(db)
	locker := &utils.BcryptLocker{}
	tokens := token.NewJWTManager(ai.JWTSecret, ai.TokenTTL)

	userSvc := user_service.NewService(userRepo, locker, tokens)
	userHandler := controller.NewUserHandler(userSvc)
	authMw := middleware.NewAuthMiddleware(tokens)

	hostingRepo := db_driver.NewHostingRepository(db)
	ctx := context.Background()

	// AWS Config 로드 (환경변수 또는 IAM 역할 기반)
	cfg, err := config.LoadDefaultConfig(ctx)
	ec2Manager := aws.NewEC2Manager(cfg)
	if err != nil {
		panic(err)
	}

	hostingSvc := hosting_service.NewService(hostingRepo, "172.31.32.87:5003", ec2Manager)
	hostingHandler := controller.NewHostingHandler(hostingSvc, userSvc)
	return &HandlerRegistry{
		UserHandler:    userHandler,
		JWTManager:     tokens,
		AuthMiddleware: authMw,
		HostingHandler: hostingHandler,
	}, nil
}

func (c DBConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		c.User, c.Password, c.Host, c.Port, c.Name)
}
