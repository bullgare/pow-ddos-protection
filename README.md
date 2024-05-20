# pow-ddos-protection

It is a home task.

## Requirements

Design and implement “Word of Wisdom” tcp server.
* TCP server should be protected from DDOS attacks with the Prof of Work, the challenge-response protocol should be used.
* The choice of the POW algorithm should be explained.
* After Prof Of Work verification, server should send one of the quotes from “word of wisdom” book or any other collection of the quotes.
* Docker file should be provided both for the server and for the client that solves the POW challenge

## Logic

* Clients first send an authorization request (requests a challenge).
* Server sends back a unique seed for the client together with a config for POW challenge, 
  * and also stores the seed and the client's data in a cache.
* Client performs some work to satisfy the rules sent by the server based on the seed.
* Client then sends the data request providing a token, original seed and rules received from the server earlier.
* Server checks the token and if it's okay, returns a quote from the Book.
  * It checks if client's identity matches the client that requested the challenge originally,
  it also checks the stored config. After checking provided token, data is removed from the cache.

Server adjusts the complexity of the challenge based on the target RPS.

## How to run

`make run-in-docker` Starts the server, its cache and 2 clients in docker. Each client will make 30 requests for a quote.

`make rebuild-and-run-in-docker` The same as above, but it also rebuilds server and clients.

### Configuration

Configuration is done through env variables (specified in [docker-compose](./build/dev/docker-compose.yaml))

#### Server

* `NETWORK_ADDRESS` - address to listen to
* `REDIS_ADDRESS` - cache address
* `TARGET_RPS` - target rps to adjust POW params
* `INFO_LOGS_ENABLED` - whether you want info level logs from the server

#### Client

* `NETWORK_ADDRESS` - address to send requests to

### Tests

`make test` is to run unit tests.

`make test-integration` is to run integration tests.


## Implementation details

This project consists of a server and a client, implemented in Golang.

Each of them has a separate `main` file, but they share the codebase.

The project is using a sort of Clean Architecture pattern.

### Server

It starts an infinite tcp listener,
and for each incoming connection starts another infinite loop for communication with a particular client.

In `./server/main.go` we instantiate all the entities:
tcp listener, repositories for cache and quotes, hashcash (used for POW), and handlers.

#### Main directories

`./internal/domain` in our case only has contracts for the repositories: auth storage and word o wisdom.

`./internal/usecase/handlers/server` has protocol-agnostic handlers

`./internal/app/server` knows how to map usecase handlers req/resp to the transport layer

`./internal/infra` holds implementations for contracts defined on upper layers:
* `./internal/infra/auth/hashcash` is an implementation of the POW engine. It is pluggable, 
can be easily replaced as it is just an implementation of the `./internal/usecase/contracts/auth` contract.
* `./internal/infra/repositories` consists of implementations of the domain contracts for repos.
* `./internal/infra/protocol` has message-specific logic
* `./internal/infra/transport` has transport-specific logic

### Client

In `./client/main.go` we instantiate all the entities:
wow client, hashcash for POW, and starts a loop for getting a wow quote for 30 times.

In each loop, it first requests `auth` endpoint, generates a token based on params taken from the response,
and then requests `data` endpoint with the generated token, and original seed.

#### Main directories

`./internal/usecase/handlers/client` - logic for each run sits here

`./internal/infra/auth/hashcash` as described in the `server` section, is an implementation of the POW engine

### Difficulty manager

`./internal/infra/auth/hashcash/difficulty_manager.go`

Any request to a wow quote is added to a bucket (as `IncrRequests` is invoked).

Any request for a challenge is leading to requesting a difficulty percentage (`GetDifficultyPercent`).

Difficulty manager has 2 buckets: the current bucket and the previous bucket.
And it has a bucket duration (5s). Every 5 seconds manager atomically switches buckets (current becomes previous, and current is being flushed).

For any `GetDifficultyPercent` request, 
it calculates the fraction of number of requests from the previous bucket multiplied by the elapsed time,
and adds it to the number of requests in the current bucket.

If that number (divided by number of seconds in a time frame) is greater than the target,
it adds a `step` to the current difficulty, if it's lower, then it deducts a `step`.

### Project structure

```sh
│
├── build
│   └── dev
│       ├── Dockerfile_client
│       ├── Dockerfile_server
│       └── docker-compose.yaml
├── cmd
│   ├── client
│   │   └── main.go
│   └── server
│       ├── config.go
│       └── main.go
├── internal
│   ├── app # starts the server based on the configs
│   │   └── server
│   ├── domain # domain layer logic (only contracts in our case)
│   │   └── contracts
│   ├── infra # infra layer; implementations of all the contracts defined on upper layers
│   │   ├── auth
│   │   │   └── hashcash # has generator itself, checker, difficulty manager, and seed generator
│   │   ├── clients # a client for our server
│   │   │   └── wordofwisdom
│   │   ├── protocol # logic related to protocol
│   │   │   └── common
│   │   ├── repositories # repos
│   │   │   ├── auth_storage.go
│   │   │   └── word_of_wisdom.go
│   │   └── transport # tcp transport specific logic
│   │       ├── client
│   │       ├── connection.go
│   │       └── listener
│   └── usecase # usecase layer
│       ├── contracts
│       │   ├── auth.go
│       │   └── communication.go
│       ├── handlers # client and server handlers
│       │   ├── client
│       │   └── server
│       │       ├── auth.go
│       │       └── data.go
│       └── users # user data
│           ├── context.go
│           └── user.go
└── tests # integration tests
    └── integration_server_test.go

```

## POW choice explanation

I was checking the following 5:
* Hashcash
* CryptoNight
* Scrypt
* Equihash
* Argon2

#### Hashcash

Classics. 

Pros: simple, easy to implement, CPU-bound, easy and cheap to check on a server side if it's valid.
Cons: could be vulnerable to ASICs, hard to predict generation time.

#### CryptoNight

Memory-bound, no good library, hard to check on a server side

#### Scrypt
Uses more memory for calculating a hash, not that easy to be brute-forced by ASICs.
Needs to be recalculated to check if it's valid.

#### Equihash 

Memory-bound, hard to parallelize.
Needs to be recalculated to check if it's valid.

#### Argon2
Memory-bound, resistant to GPU and ASIC attacks.
Needs to be recalculated to check if it's valid.

### Decision

I've decided to continue with the Hashcash as it's pretty simple, does not consume much server resources to check the validity.

And can be replaced with something else later if needed.

## Things to improve

* add ci/cd
* add more unit-tests
* add a linter
* add load tests
* use parametrized errors for better error segregation (not exposing internals to clients as it is now)

## Excuses

There are a lot of dependencies in `go.mod`, but the majority of them are related to integration tests.
