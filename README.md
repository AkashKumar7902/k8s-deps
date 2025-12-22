# Go API K8s Test Service

This project acts as a testbed for Kubernetes configurations, providing a Go-based API service that interacts with MongoDB and PostgreSQL. It includes various endpoints for testing purposes, such as delays, random data generation, and external API calls.

## Prerequisites

Before running this project, ensure you have the following installed:

*   **Docker**: For building the container image.
*   **Kind (Kubernetes in Docker)**: For running a local Kubernetes cluster.
*   **Kubectl**: For interacting with the Kubernetes cluster.

## Setup & Deployment

Follow these steps to deploy the application to your local Kind cluster.

### 1. Build the Docker Image
Build the Go application container image.
```bash
docker build -t my-go-server:latest .
```

### 2. Load Image into Kind
Load the built image into your Kind cluster so it's accessible to the nodes.
```bash
kind load docker-image my-go-server:latest
```

### 3. Deploy to Kubernetes
Apply the Kubernetes manifests to create the deployment and service.
```bash
kubectl apply -f k8s
```

## Access the Service

### Port Forwarding
Since the service runs inside the cluster, you need to port-forward it to access the API locally.
```bash
kubectl port-forward svc/go-api-service 8080:80
```

### API Endpoints
Once port-forwarding is active, you can interact with the service using the following commands:

#### **Basic Endpoints**
*   **Hello World** - Simple connectivity check.
    ```bash
    curl http://localhost:8080/hello
    ```
*   **Noisy** - Returns random data (useful for testing logs/response parsing).
    ```bash
    curl http://localhost:8080/noisy
    ```
*   **Slow** - Simulates a slow response (35s delay).
    ```bash
    curl http://localhost:8080/slow
    ```

#### **Database Integrations**
*   **Mongo DB** - Connects to the MongoDB service.
    ```bash
    curl http://localhost:8080/mongo
    ```
*   **Postgres DB** - Connects to the PostgreSQL service.
    ```bash
    curl http://localhost:8080/postgres
    ```

#### **External Calls**
*   **External API** - Makes an outbound call to a public API.
    ```bash
    curl http://localhost:8080/external
    ```