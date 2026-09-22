# MBA Golang expert 03 - Clean Arch Challenge 

Ao avaliador do desafio, esse repositorio serve tambem serve como respositorio de aprendizado entao terao mais informacoes que o necessario pra avaliacao

Pra voce basta apenas rodar o comando pra subir o docker compose

```bash
docker-compose up -d
```

ou com o Makefile presente

```bash
make docker
```

Nele vai rodar os container do mysql, da migration, do rabbit mq e o app e vai expor as seguintes portas: 8000 webserver rest), 8080 (graphql) e 50051 (rabbitmq)

## Ferramentas

Esse projeto usa graphql, grpc e rest entao pra fins de debugg voce precisa das seguintes ferramentas

### Rest

Com a extensao do vs code rest client temos duas rotas na pasta `api/`, nossa rota esta presente em `list_orders.http`, ou rodar um curl `GET` pra rota `http://localhost:8000/orders HTTP/1.1`

### Graphql

Nesse projeto estamos usando a tool do gqlgen, cabe olhar a [documentacao](https://gqlgen.com/) pra ver se mudou mas tem fazer o get da ferramente no projeto com

```
go get -tool github.com/99designs/gqlgen
```

Pra usar o playground graphql usamos ela em 

```
localhost:8080
```

### gRPC

Como ferramenta de chamada pro grpc usamos o [evans](https://github.com/ktr0731/evans) pra isso chamamos o service presente em OrderService e nossa chamada e feita com

```rpc
call ListOrders
```

### Migrations

Para as migrations estamos usando o `sqlc` e o pacote [`golang-migrate`](https://github.com/golang-migrate/migrate)

pra criar uma nova migrate rode com `migrate -crete ext=sql -dir=sql/migrations -seq <nome>` ou `make migrate`

### Dependency Injection

Pra isso usamos o wire, se adicionar um novo use case adicione essa dependencia de injecao no `cmd/ordersystem/wire.go` e rode `make wire`

### Docker 

Banco de dados mysql e servicos rodam com docker compose com o comando `make docker`

## Arquitetura

Esse projeto segue a Clean Architecture, organizando o codigo em camadas concentricas onde as camadas mais internas nao conhecem as mais externas.

![Clean Architecture](docs/images/clean-architecture.jpg)

| Camada | Circulo na imagem | Pasta no projeto | Conteudo |
|---|---|---|---|
| Entities | Amarelo (centro) | `internal/entity` | `Order`, `OrderRepositoryInterface` - as regras de negocio mais genericas, sem dependencia de nada externo |
| Use Cases | Vermelho | `internal/usecase` | `CreateOrderUseCase`, `ListOrdersUseCase` - orquestram as entidades para cumprir um caso de uso especifico da aplicacao |
| Interface Adapters (Controllers / Presenters / Gateways) | Verde | `internal/infra/web` (controller REST), `internal/infra/grpc/service` (controller gRPC), `internal/infra/graph` (controller/presenter GraphQL), `internal/infra/database` (gateway - `OrderRepository` implementando o repository), `internal/event` e `internal/event/handler` (presenter que publica o evento `OrderCreated` no RabbitMQ) | Adaptam dados entre os use cases e o mundo externo (HTTP, gRPC, GraphQL, banco, fila) |
| Frameworks & Drivers (Web / DB / Devices / External Interfaces) | Azul (borda externa) | `cmd/ordersystem` (entrypoint, wiring com `wire`), `configs` (viper), `internal/infra/web/webserver` (chi), `internal/infra/grpc/pb` (protobuf/grpc gerado), `sql/migrations` (golang-migrate), driver `go-sql-driver/mysql`, cliente `streadway/amqp` (RabbitMQ) | Frameworks, drivers e ferramentas concretas plugadas nos adapters |

O fluxo de controle segue a seta rosa da imagem: um Controller (ex.: `WebOrderHandler.Create`) chama o Use Case Interactor (`CreateOrderUseCase.Execute`), que usa as Entities e devolve a saida via Output Port/DTO para um Presenter (ex.: o handler que publica `OrderCreated` no RabbitMQ ou o resolver GraphQL que formata a resposta).
