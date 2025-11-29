package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"shortener/src/internal/application/config"
	"shortener/src/internal/application/contracts"
	"shortener/src/internal/application/services"
	shortlink "shortener/src/internal/domain/short_link"
	"shortener/src/internal/domain/visit"
	"shortener/src/internal/infrastructure/data"
	"shortener/src/internal/infrastructure/data/repositories"
	"shortener/src/internal/infrastructure/kafka"
	generator "shortener/src/internal/infrastructure/short_link_generator"
	"shortener/src/internal/web_api/controllers"
	"shortener/src/pkg/logger"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/dbpg"
	wbfkafka "github.com/wb-go/wbf/kafka"
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

	visitService, shortLinkService := initServices(visitRepository, shortLinkRepository, producer, codeGenerator)

	validate := validator.New()

	shortLinkController := initControllers(shortLinkService, visitService, validate)

	server := initServer(cfg.HTTP, shortLinkController)
	go func() {
		err := server.ListenAndServe()
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				log.Fatal(err)
			}
		}
	}()

	kafkaConsumer := wbfkafka.NewConsumer([]string{cfg.Kafka.Broker}, cfg.Kafka.Topic, cfg.Kafka.GroupID)
	consumer := initVisitConsumer(kafkaConsumer, visitService, retry.Strategy{
		Attempts: 3,
		Delay:    time.Duration(0.5 * float64(time.Second)),
		Backoff:  2,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	consumer.Start(ctx)

	gracefulShutdown(cancel, server, consumer)
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
) (visit.VisitService, shortlink.ShortLinkService) {
	return services.NewVisitService(visitRepository, producer),
		services.NewShortLinkService(shortLinkRepository, generator)
}

func initControllers(
	shortLinkService shortlink.ShortLinkService,
	visitService visit.VisitService,
	validator *validator.Validate,
) *controllers.ShortLinkController {
	return controllers.NewShortLinkController(shortLinkService, visitService, validator)
}

func initServer(cfg config.HTTPConfig, shortLinkController *controllers.ShortLinkController) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	shortLinkController.UseHandlers(r)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	return server
}

func gracefulShutdown(cancelFunc context.CancelFunc, server *http.Server, consumer *kafka.VisitConsumer) {
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
				logger.Error("failed to gracefully shutdown the server", err)
			}
		}
	})

	wg.Go(func() {
		if err := consumer.Stop(); err != nil {
			logger.Error("failed to gracefully shutdown the consumer", err)
		}
	})

	wg.Wait()
}
