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

## What's implemented

A simple PBFT consensus with VRF proposer selection. Time is divided into rounds and each round is subdivided into 
a proposal phase and an agreement phase. We implement the VRF-based proposing in the proposal phase, and we impelement
a simple proposal reduction that executes immediately after the proposal phase ends. From there the winning proposal is
handed to the PBFT-like agreement phase. We implement a simple key-value store on top of our consensus protocol to showcase 
it's usage. Because our protocol does not reply to the client directly after commiting and executing, we also implement a 
client that queries all nodes and only trust the values that are the same on f+1 nodes.

## Who did what

Adarsh: PBFT related code: service states, message handlers for processing incoming Proposal, Prepare, and Commit messages
Anjali: PBFT related code: service states, processor functions for proposing, preparing, committing and executing proposals.
Patrick: overall structure of the program, peripherals (tests and docker setups), VRF proposing, network layer implementation.
