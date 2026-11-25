# Smart Garden Bot - Project Overview

## Purpose
Smart Garden Bot is a scalable SaaS platform that automates garden watering through intelligent weather-based decision making and IoT sensor integration.

## Architecture
- **Web Application**: Mobile-responsive Next.js frontend with TypeScript
- **REST API**: Go-based microservice with OpenAPI 3.1 specification  
- **Kubernetes Operator**: Go-based controller for IoT device management
- **Database**: PostgreSQL with pgvector for AI/ML capabilities
- **Infrastructure**: Cloud-native deployment with Kubernetes and Helm

## Tech Stack
- **Backend**: Go 1.18+ with Gin framework
- **Frontend**: Next.js 15.4+ with React 18.3+, TypeScript 5.7+
- **Database**: PostgreSQL with CloudNativePG operator
- **Container Orchestration**: Kubernetes with Kind for local development
- **Development Tools**: Tilt for hot-reloading, Task for command runner
- **Testing**: Go testify, Jest for frontend, Cypress for E2E
- **State Management**: Zustand for frontend
- **Styling**: Tailwind CSS with Radix UI components

## Development Environment
- **Local Setup**: Kind + Tilt with hot-reloading
- **Task Runner**: Taskfile.yml for standardized commands
- **Container Runtime**: Docker Desktop 4.20+
- **Prerequisites**: Node.js 24+, Go 1.18+, kubectl 1.28+, Kind 0.20+, Tilt 0.33+