package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"time"

	graphql_handler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/hiaguedes/mba-go-challenge-03-clean-arch/configs"
	"github.com/hiaguedes/mba-go-challenge-03-clean-arch/internal/event/handler"
	"github.com/hiaguedes/mba-go-challenge-03-clean-arch/internal/infra/graph"
	"github.com/hiaguedes/mba-go-challenge-03-clean-arch/internal/infra/grpc/pb"
	"github.com/hiaguedes/mba-go-challenge-03-clean-arch/internal/infra/grpc/service"
	"github.com/hiaguedes/mba-go-challenge-03-clean-arch/internal/infra/web/webserver"
	"github.com/hiaguedes/mba-go-challenge-03-clean-arch/pkg/events"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	// mysql
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	configs, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open(configs.DBDriver, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", configs.DBUser, configs.DBPassword, configs.DBHost, configs.DBPort, configs.DBName))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	waitForDB(db)

	rabbitMQChannel := getRabbitMQChannel(configs.RabbitMQUser, configs.RabbitMQPassword, configs.RabbitMQHost, configs.RabbitMQPort)

	eventDispatcher := events.NewEventDispatcher()
	eventDispatcher.Register("OrderCreated", &handler.OrderCreatedHandler{
		RabbitMQChannel: rabbitMQChannel,
	})

	createOrderUseCase := NewCreateOrderUseCase(db, eventDispatcher)
	listOrdersUseCase := NewListOrdersUseCase(db)

	webserver := webserver.NewWebServer(configs.WebServerPort)
	webOrderHandler := NewWebOrderHandler(db, eventDispatcher)
	webserver.AddHandler("/order", webOrderHandler.Create)
	webserver.AddHandler("/orders", webOrderHandler.List)
	fmt.Println("Starting web server on port", configs.WebServerPort)
	go webserver.Start()

	grpcServer := grpc.NewServer()
	createOrderService := service.NewOrderService(*createOrderUseCase, *listOrdersUseCase)
	pb.RegisterOrderServiceServer(grpcServer, createOrderService)
	reflection.Register(grpcServer)


	fmt.Println("Starting gRPC server on port", configs.GRPCServerPort)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", configs.GRPCServerPort))
	if err != nil {
		panic(err)
	}
	go grpcServer.Serve(lis)

	srv := graphql_handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &graph.Resolver{
		CreateOrderUseCase: *createOrderUseCase,
		ListOrdersUseCase:  *listOrdersUseCase,
	}}))
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	fmt.Println("Starting GraphQL server on port", configs.GraphQLServerPort)
	http.ListenAndServe(":"+configs.GraphQLServerPort, nil)
}

func waitForDB(db *sql.DB) {
	for i := 0; i < 30; i++ {
		if err := db.Ping(); err == nil {
			return
		}
		time.Sleep(2 * time.Second)
	}
	panic("could not connect to database after retries")
}

func getRabbitMQChannel(user, password, host, port string) *amqp.Channel {
	dsn := fmt.Sprintf("amqp://%s:%s@%s:%s/", user, password, host, port)

	var conn *amqp.Connection
	var err error
	for i := 0; i < 30; i++ {
		conn, err = amqp.Dial(dsn)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	return ch
}
