package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	companymysql "github.com/ktsiligkos/xm_project/internal/repository/company/mysql"
	companyservice "github.com/ktsiligkos/xm_project/internal/service/company"

	kafkaevents "github.com/ktsiligkos/xm_project/internal/platform/events/kafka"
	usermysql "github.com/ktsiligkos/xm_project/internal/repository/user/mysql"
	userservice "github.com/ktsiligkos/xm_project/internal/service/user"
	httptransport "github.com/ktsiligkos/xm_project/internal/transport/http"
	"github.com/ktsiligkos/xm_project/pkg/config"
)

// Application owns the assembled HTTP server and its dependencies.
type Application struct {
	engine           *gin.Engine
	cfg              config.Config
	db               *sql.DB
	logger           *zap.Logger
	companyPublisher *kafkaevents.Publisher
}

// New wires dependencies together and prepares the HTTP server.
func New(cfg config.Config) (*Application, error) {

	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("init zap logger: %w", err)
	}

	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		log.Printf("mysql setup failed, using in-memory repository: %v", err)
	}

	// wire the company service
	companyRepo := companymysql.NewMySQL(db)
	eventPublisher := kafkaevents.NewPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	companyService := companyservice.NewService(companyRepo, eventPublisher)
	companiesHandler := httptransport.NewCompaniesHandler(companyService, logger.Named("companies_handler"))

	// wire the user service
	userRepo := usermysql.NewMySQL(db)
	userService := userservice.NewService(userRepo, []byte(cfg.JWTSecret), time.Hour*1)
	usersHandler := httptransport.NewUsersHandler(userService, logger.Named("users_handler"))

	router := httptransport.NewRouter(companiesHandler, usersHandler, []byte(cfg.JWTSecret))

	return &Application{
		engine:           router,
		cfg:              cfg,
		db:               db,
		logger:           logger,
		companyPublisher: eventPublisher,
	}, nil
}

// Run starts the HTTP server.
func (a *Application) Run() error {
	return a.engine.Run(a.cfg.HTTPAddr)
}

// Handler exposes the underlying HTTP handler for tests.
func (a *Application) Handler() http.Handler {
	return a.engine
}

// Close releases any resources owned by the application.
func (a *Application) Close() error {
	var firstErr error

	if a.companyPublisher != nil {
		if err := a.companyPublisher.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if a.logger != nil {
		_ = a.logger.Sync()
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
