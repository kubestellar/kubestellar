
# KubestellarUI Setup Guide

Welcome to **KubestellarUI**! This guide will help you set up the KubestellarUI application on your local machine after cloning the repository for development. The application consists of two main parts:

1. **Frontend**: Built with React and TypeScript
2. **Backend**: Built with Golang using the Gin framework.

## Contents

- [Prerequisites](#prerequisites)
- [Installation Steps](#installation-steps)
  - [Local Setup](#local-setup)
  - [Local Setup with Docker Compose](#local-setup-with-docker-compose)
- [Docker Image Versioning and Pulling](#docker-image-versioning-and-pulling)
- [Accessing the Application](#accessing-the-application)

## Prerequisites

Before you begin, ensure that your system meets the following requirements:

### 1. Golang

- **Version**: 1.23.4
- **Download Link**: [Golang Downloads](https://golang.org/dl/)

### 2. Node.js and npm

- **Node.js Version**: ≥ 16.x.x
- **npm Version**: Comes bundled with Node.js
- **Download Link**: [Node.js Downloads](https://nodejs.org/en/download/)

> [!NOTE]
> You can use [nvm](https://github.com/nvm-sh/nvm) to manage multiple Node.js versions.

### 3. Git

- Ensure Git is installed to clone the repository
- **Download Link**: [Git Downloads](https://git-scm.com/downloads)

### 4. Kubernetes Clusters

- Ensure you have access to a Kubernetes clusters setup with Kubestellar Getting Started Guide & Kubestellar prerequisites installed

- **Kubestellar guide**: [Guide](https://docs.kubestellar.io/release-0.25.1/direct/get-started/)

## Installation Steps

Clone the Repository

```bash
git clone https://github.com/your-github-username/ui.git

cd ui
```
Then go through one of the setup options below:
- [Local Setup](#local-setup)
- [Local Setup with Docker Compose](#local-setup-with-docker-compose)

### Local Setup

#### Step 1: Create `.env` File for Frontend Configuration

To configure the frontend, copy the `example.env` file to a `.env` file in the project root directory (where `package.json` is located).

```bash
cp example.env .env
```

**Example `.env` file:**  

```
VITE_BASE_URL=http://localhost:4000
VITE_APP_VERSION=0.1.0
VITE_GIT_COMMIT_HASH=$GIT_COMMIT_HASH
```

> [!NOTE] 
> This is because `.env` files are intended to be a personal environment configuration file. The included `example.env` in the repo is a standard that most other node projects include for the same purpose. You rename the file to `.env` and then change its contents to align with your system and personal needs.

##### Tracking Application Version and Git Commit Hash

KubestellarUI uses environment variables to track the app version and the current Git commit hash.  

**Environment Variables**  

| Variable               | Purpose                                 | Example |
|------------------------|-----------------------------------------|---------|
| `VITE_BASE_URL`        | Defines the base URL for API calls     | `http://localhost:4000` |
| `VITE_APP_VERSION`     | Defines the current application version | `0.1.0` |
| `VITE_GIT_COMMIT_HASH` | Captures the current Git commit hash   | (Set during build) |


#### Step 2: Run Redis Container (Optional)

KubestellarUI uses Redis for caching real-time WebSocket updates to prevent excessive Kubernetes API calls.  

Run Redis using Docker:  

```bash
docker run --name redis -d -p 6379:6379 redis
```

Verify Redis is running:  

```bash
docker ps | grep redis
```

#### Step 3: Install and Run the Backend

Make sure you are in the root directory of the project

```bash
cd backend

go mod download

go run main.go
```

You should see output indicating the server is running on port `4000`.

#### Step 4: Install and Run Frontend

Open another terminal and make sure you are in the root directory of the project.

```bash
npm install

npm run dev
```

You should see output indicating the server is running on port `5173`.

### Local Setup with Docker Compose

If you prefer to run the application using Docker Compose, follow these steps:

#### Step 1: Ensure Docker is Installed

- **Download Link**: [Docker Downloads](https://www.docker.com/products/docker-desktop)

> [!NOTE] 
> If you are using Compose V1, change the `docker compose` command to `docker-compose` in the following steps.
> Checkout [Migrating to Compose V2](https://docs.docker.com/compose/releases/migrate/) for more info.

#### Step 2: Run Services

From the project root directory

```bash
docker compose up --build
```

You should see output indicating the services are running.

To stop the application

```bash
docker compose down
```

#### Use Docker Compose in Development Cycle

For ongoing development, use the following steps:

- **Step 1: Stop the running Application**:
  ```bash
  docker compose down
  ```

- **Step 2: Pull the Latest Source Code Changes**:
  ```bash
  git pull origin main
  ```

- **Step 3: Rebuild and Restart the Application**:
  ```bash
  docker compose up --build
  ```
This will:

- Stop the running containers.
- Pull the latest source code changes.
- Rebuild and restart the application.

## **🚀 Install GolangCI-Lint**

To install **GolangCI-Lint**, follow these steps:

### **🔹 Linux & macOS**
Run the following command:
```sh
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
```
Ensure `$(go env GOPATH)/bin` is in your `PATH`:
```sh
export PATH=$(go env GOPATH)/bin:$PATH
```

### **🔹 Windows**
Use **scoop** (recommended):
```powershell
scoop install golangci-lint
```
Or **Go install**:
```sh
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### **🔹 Verify Installation**
Run:
```sh
golangci-lint --version
```

---

## **🛠 Linting & Fixing Code**
### **🔹 Check for Issues**
```sh
make check-lint
```
### **🔹 Auto-Fix Issues**
```sh
make fix-lint
```
### **🔹 Run Both**
```sh
make lint
```

### Docker Image Versioning and Pulling

If you'd like to work with the Docker images for the **KubestellarUI** project, here's how you can use the `latest` and versioned tags:

1. **Frontend Image**:
   - Tag: `quay.io/kubestellar/ui:frontend`
   - Latest Version: `latest`
   - Specific Version (Commit Hash): `frontend-<commit-hash>`

2. **Backend Image**:
   - Tag: `quay.io/kubestellar/ui:backend`
   - Latest Version: `latest`
   - Specific Version (Commit Hash): `backend-<commit-hash>`

#### How to Pull the Latest Images:

- **Frontend Image**:
  ```bash
  docker pull quay.io/kubestellar/ui:frontend
  ```

- **Backend Image**:
  ```bash
  docker pull quay.io/kubestellar/ui:backend
  ```

#### How to Pull Specific Version (Commit Hash):

If you want to pull an image for a specific version (e.g., commit hash), use:

- **Frontend Image with Version**:
  ```bash
  docker pull quay.io/kubestellar/ui:frontend-abcd1234
  ```

- **Backend Image with Version**:
  ```bash
  docker pull quay.io/kubestellar/ui:backend-abcd1234
  ```


### Accessing the Application

1. **Backend API**: [http://localhost:4000](http://localhost:4000)
2. **Frontend UI**: [http://localhost:5173](http://localhost:5173)


<div>
<h2><font size="6"><img src="https://raw.githubusercontent.com/Tarikul-Islam-Anik/Animated-Fluent-Emojis/master/Emojis/Smilies/Red%20Heart.png" alt="Red Heart" width="40" height="40" /> Contributors </font></h2>
</div>
<br>

<center>
<a href="https://github.com/kubestellar/ui/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=kubestellar/ui" />
</a>
</center>
<br>
