// @title			Shortener API
// @version		1.0
// @description	Сервис для создания коротких ссылок и получения аналитики по переходам.
// @BasePath		/
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortener/src/internal/application/config"
	"shortener/src/internal/application/contracts"
	"shortener/src/internal/application/services"
	shortlink "shortener/src/internal/domain/short_link"
	"shortener/src/internal/domain/visit"
	"shortener/src/internal/infrastructure/cache"
	"shortener/src/internal/infrastructure/data"
	"shortener/src/internal/infrastructure/data/repositories"
	"shortener/src/internal/infrastructure/kafka"
	generator "shortener/src/internal/infrastructure/short_link_generator"
	"shortener/src/internal/web_api/controllers"
	"shortener/src/internal/web_api/public"
	"shortener/src/pkg/logger"
	"sync"
	"syscall"
	"time"

	_ "shortener/src/internal/web_api/docs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	httpSwagger "github.com/swaggo/http-swagger"
	"github.com/wb-go/wbf/dbpg"
	wbfkafka "github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/retry"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	logger.Init(cfg.LogLevel)

	db, err := data.InitDb(cfg.Postgres)
	if err != nil {
		log.Fatal(err)
	}

	visitRepository, shortLinkRepository := initRepositories(db, retry.Strategy{
		Attempts: 3,
		Delay:    time.Duration(0.5 * float64(time.Second)),
		Backoff:  2,
	})

	kafkaProducer := wbfkafka.NewProducer([]string{cfg.Kafka.Broker}, cfg.Kafka.Topic)
	producer := initKafkaProducer(kafkaProducer, retry.Strategy{
		Attempts: 3,
		Delay:    time.Duration(0.5 * float64(time.Second)),
		Backoff:  2,
	})

	codeGenerator := generator.NewRandomShortCodeGenerator()

	redisClient := redis.New(cfg.Redis.Host, cfg.Redis.Password, cfg.Redis.DB)
	redisCache := cache.NewRedis(redisClient, retry.Strategy{
		Attempts: 3,
		Delay:    time.Duration(0.5 * float64(time.Second)),
		Backoff:  2,
	})

	visitService, shortLinkService := initServices(
		visitRepository,
		shortLinkRepository,
		producer,
		codeGenerator,
		redisCache,
	)

	validate := validator.New()

	shortLinkController, analyticsController := initControllers(shortLinkService, visitService, validate)

	server := initServer(cfg.HTTP, shortLinkController, analyticsController)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}
	}()
	logger.Info(fmt.Sprintf("shortener server listening on port %s", cfg.HTTP.Port))

	kafkaConsumer := wbfkafka.NewConsumer([]string{cfg.Kafka.Broker}, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	logger.Info(cfg.Kafka.GroupID)
	consumer := initVisitConsumer(kafkaConsumer, visitService, retry.Strategy{
		Attempts: 3,
		Delay:    time.Duration(0.5 * float64(time.Second)),
		Backoff:  2,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)
	logger.Info("visit consumer started")

	gracefulShutdown(cancel, server, consumer, redisClient, db)
}

func initRepositories(db *dbpg.DB, retry retry.Strategy) (visit.VisitRepository, shortlink.ShortLinkRepository) {
	return repositories.NewVisitRepository(db, retry), repositories.NewShortLinkRepository(db, retry)
}

func initVisitConsumer(
	consumer *wbfkafka.Consumer,
	visitService visit.VisitService,
	consumerRetry retry.Strategy,
) *kafka.VisitConsumer {
	return kafka.NewVisitConsumer(consumer, visitService, consumerRetry)
}

func initKafkaProducer(producer *wbfkafka.Producer, producerRetry retry.Strategy) contracts.MessageProducer {
	return kafka.NewKafkaProducer(producer, producerRetry)
}

func initServices(
	visitRepository visit.VisitRepository,
	shortLinkRepository shortlink.ShortLinkRepository,
	producer contracts.MessageProducer,
	generator shortlink.ShortLinkGenerator,
	redis contracts.Cache,
) (visit.VisitService, shortlink.ShortLinkService) {
	return services.NewVisitService(visitRepository, producer),
		services.NewShortLinkService(shortLinkRepository, generator, redis)
}

func initControllers(
	shortLinkService shortlink.ShortLinkService,
	visitService visit.VisitService,
	validator *validator.Validate,
) (*controllers.ShortLinkController, *controllers.AnalyticsController) {
	return controllers.NewShortLinkController(shortLinkService, visitService, validator),
		controllers.NewAnalyticsController(visitService)
}

func initServer(
	cfg config.HTTPConfig,
	shortLinkController *controllers.ShortLinkController,
	analyticsController *controllers.AnalyticsController,
) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	shortLinkController.UseHandlers(r)
	analyticsController.UseHandlers(r)
	public.UseStaticFiles(r)

	r.Get("/swagger/*", httpSwagger.WrapHandler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	return server
}

func gracefulShutdown(
	cancelFunc context.CancelFunc,
	server *http.Server,
	consumer *kafka.VisitConsumer,
	redisClient *redis.Client,
	db *dbpg.DB,
) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	<-c
	cancelFunc()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	wg := &sync.WaitGroup{}

	wg.Go(func() {
		if err := server.Shutdown(ctx); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Error("failed to gracefully shutdown the server", "err", err)
			}
		}
		logger.Info("shortener server gracefully closed")

		if err := redisClient.Close(); err != nil {
			logger.Error("failed to close redis client", "err", err)
		}
		logger.Info("redis client closed")

		if err := db.Master.Close(); err != nil {
			logger.Error("failed to close db", "err", err)
		}

		for _, slave := range db.Slaves {
			if err := slave.Close(); err != nil {
				logger.Error("failed to close slave db")
			}
		}
		logger.Info("db closed")
	})

	wg.Go(func() {
		if err := consumer.Stop(); err != nil {
			logger.Error("failed to close consumer", "err", err)
		}

		logger.Info("consumer stopped")
	})

	wg.Wait()
}
