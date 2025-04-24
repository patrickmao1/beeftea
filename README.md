# beeftea

## Set up project

1. Make sure you have `go` installed

2. Install protobuf toolchain (requires MacOS and homebrew)

```shell
brew install protobuf
```

3. Install the [gRPC plugin](https://grpc.io/docs/languages/go/quickstart/)

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

## Compile proto files to go types

Make sure you have the `make` binary to use `Makefile`

```shell
make proto
```

## Running a local cluster

We use docker to test a cluster of nodes. Make sure you have docker installed.

The `compose.yaml` file in the project root defines a local network of five beeftea nodes, each with a fixed custom 
private IP.

```shell
# builds the beeftea docker image according to the Dockerfile in the project root, then bring up the cluster defined 
# in compose.yaml.
docker compose up --build -d
```

Logs produced by running the cluster will be located at `./tests/beeftea/docker_volumes/node[X]`