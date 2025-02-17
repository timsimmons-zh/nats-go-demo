# NATS Go Demo
This project demonstrates the setup of a NATS cluster with three nodes and multiple clients (producers and consumers) using Docker Compose. The docker-compose.yaml file defines the services and configurations required to establish the cluster and run the clients.

## Services
### NATS Cluster
The NATS cluster consists of three nodes:

n1.local.dev
n2.local.dev
n3.local.dev
Each node is configured to join the cluster and store data in named volumes.

### Clients
The clients include:

nats-client-1
nats-client-2
nats-client-3
These clients are configured to act as producers and consumers, interacting with the NATS cluster.

## Profiles
The `docker-compose.yaml` file uses two profiles:

- *nats*: Establishes the NATS cluster.
- *client*: Spins up the producer and consumer clients.

### Usage
Start the NATS cluster using the nats profile:

```bash
docker compose --profile nats up -d 
```
Wait for the leader election to complete. You can check the logs to ensure the cluster is up and running:
```bash
docker compose --profile nats logs | grep "JetStream cluster new metadata leader"
```

Once the NATS cluster is ready, start the clients using the client profile:
```bash
docker-compose --profile client up --build
```

### Tearing Down the Cluster
The NATS servers use named volumes to persist the stream state. To fully destroy a cluster, you must remove these named volumes:
```bash
docker-compose down -v
```
