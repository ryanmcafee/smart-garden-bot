# Overview

Smart Garden Bot is an intelligent irrigation management platform comprising a mobile-responsive web application, REST API, and Kubernetes operator that automates garden watering based on weather forecasts and sensor data. Designed for tech-savvy gardeners with busy schedules, it prevents over/under-watering while optimizing plant yields through integration with weather stations and watering con

The smart-garden-bot comprises a few things: a web application, API and a kubernetes operator.

The web application allows a user to centrally manage their garden though a cross platform mobile responsive application and allows them to register their watering controllers and weather station in the app.
The API interfaces with the web application and requires OAuth2 to access. The api supports working with the web app and for standalone requests that are authenticated via an api token.
The API interfaces with a Kubernetes operator that handles controlling the watering controllers through the Kubernetes controller implementation and configured user settings.

# Goals

## Business Goals

    Create a scalable SaaS platform for automated garden management
    Generate recurring revenue through subscription-based billing
    Build a cloud-native, enterprise-grade solution with Kubernetes orchestration
    Establish integrations with leading weather data providers and IoT hardware
    Help users with planning their garden for their climate zone (what to plant and at what time)
    Help users with planning succession planting in their garden.

## User Goals

    Automate garden watering to prevent crop damage from improper irrigation
    Monitor garden conditions remotely through comprehensive sensor dashboards
    Optimize plant yields through data-driven watering decisions
    Save time by eliminating manual watering schedules and monitoring
    Advanced AI/ML plant disease detection
    Help users with planning their garden for their climate zone (what to plant and at what time)
    Help users with planning succession planting in their garden.


# Requirements

## Functional Requirements:

    Allows user to enter their zip code to get weather data specific to their location.
    Integration with weather APIs to discover future 10 day forecasts.
    - OpenWeatherMap - openweathermap.org
    - Weather.gov (NOAA)
    - https://www.weatherapi.com
    - https://open-meteo.com/
    - https://www.wunderground.com/

    User is able to register and sign in to the application using compliant OpenId Connect Providers (KeyCloak by default)

    User is able to configure a user profile to configure their address, email, phone number, and other information.

    User is able to configure their watering schedule.

    User is able to configure integration with Ecowitt weather station.
    User is able to configure integration with Ecowitt WittFlow (will allow the smart-garden-bot controller to be able to start/stop watering of zones)

    User is able to monitor sensor data with at a glance widgets (soil moisture, temperature, humidity, solar/uvi, rainfall piezo, wind, barometic pressure)
    User is able to monitor timeseries data for (soil moisture, temperature, humidity, solar/uvi, rainfall piezo, wind, barometic pressure)

    User is able to create api keys.

    User is able to configure billing information with stripe.

    User is able to monitor watering status (water is being applied, water is not being applied, etc.)

    User is able to monitor sensor data (soil moisture, temperature, humidity, etc.)

    Advanced AI/ML plant disease detection


## Non functional Requirements:

    Kubernetes operator written in Golang.
    Website created using Next.js, Deno and Typescript that showcases product, features, and pricing.
    Linting and other front end best practices are implemented.
    Authentication and authorization is handled through integration with Auth0.
    User data, preferences, etc. are stored in a Postgres database.
    Database is persisted in a postgres database.
    Postgres database is deployed in a kubernetes cluster using CloudNativePG.
    Postgres database supports pgvector to store embeddings of data to support similarity search and feature integration with an MCP server.
    REST API written in Golang that serves up a OpenAPI 3.1 spec, swagger UI, and a liveness/readiness probe endpoint.
    
    REST API has a /metrics endpoint that exposes prometheus metrics.
    REST API has a /health endpoint that returns 200 if the API is healthy.
    REST API has a /ready endpoint that returns 200 if the API is ready to serve requests.
    REST API has a /openapi.json endpoint that returns the OpenAPI 3.1 spec.
    REST API has a /swagger-ui endpoint that returns the swagger UI.
    REST API has a /liveness endpoint that returns 200 if the API is alive.
    REST API has a /readiness endpoint that returns 200 if the API is ready to serve requests.
    GitOps first deployment architecture - Published as a helm chart
    Github Actions Workflow that validates, validates and publishes artifacts (multi-platform container image + helm chart + website)






