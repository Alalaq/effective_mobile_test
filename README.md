# Effective_mobile_test

A Go application that receives full names from open APIs, enriches the data with age, gender, and nationality, and stores it in a PostgreSQL database. It also provides REST and GraphQL APIs for querying, adding, updating, and deleting person data. Additionally, data caching is implemented using Redis.

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [Project Tasks](#project-tasks)
- [Getting Started](#getting-started)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [GraphQL](#graphql)
- [Caching](#caching)
- [Testing](#testing)
- [Environment Variables](#environment-variables)
- [Contributing](#contributing)
- [License](#license)

## Project Overview

This Go application listens to a Kafka queue (`FIO`) for incoming full names in the following format:

A Go application that receives full names from open APIs, enriches the data with age, gender, and nationality, and stores it in a PostgreSQL database. It also provides REST and GraphQL APIs for querying, adding, updating, and deleting person data. Additionally, data caching is implemented using Redis.

## Table of Contents

- [Project Overview](#project-overview)
- [Architecture](#architecture)
- [Project Tasks](#project-tasks)
- [Environment Variables](#environment-variables)

## Project Overview

This Go application listens to a Kafka queue (`FIO`) for incoming full names in the following format:

```json
{
  "name": "Dmitriy",
  "surname": "Ushakov",
  "patronymic": "Vasilevich" // optional
}
```
It enriches the data with age, gender, and nationality using external APIs and saves it in a PostgreSQL database. The application exposes REST and GraphQL APIs to query, add, update, and delete person data. Data caching is implemented using Redis.

## Architecture

The project follows a structured architecture with the following components:

- application.go: The main entry point of the application, including Kafka message processing, HTTP server setup, and routing.
- entities/Person.go: Defines the Person struct to represent person data.
- service/PersonService.go: Contains the business logic for handling person data, including CRUD operations and data enrichment.
- repository/PersonRepository.go: Defines the repository interface for interacting with the PostgreSQL database.
- repository/PersonRepositoryImpl.go: Implements the PersonRepository interface and handles database operations.
- GraphQL: Defines GraphQL types and queries for interacting with the application using GraphQL.

## Project Tasks
The project implements the following tasks as specified:

1. Listens to a Kafka queue (FIO) for incoming full names and processes them.
2. Enriches correct messages with age, gender, and nationality and saves them in the PostgreSQL database.
3. Sends incorrect messages (missing mandatory fields, incorrect format) to the FIO_FAILED Kafka queue.
4. Exposes REST endpoints for various operations (GET, POST, PUT, DELETE).
5. Implements GraphQL queries and mutations.
6. Provides data caching in Redis.
7. Covers the code with logs.
8. Implements unit tests.
9. Moves configuration data to a .env file.



## Environment Variables
List of environment variables used in the project:

- 'DB_HOST': PostgreSQL database host.
- 'DB_PORT': PostgreSQL database port.
- 'DB_USER': PostgreSQL database user.
- 'DB_PASSWORD': PostgreSQL database password.
- 'DB_NAM'E: PostgreSQL database name.
- 'REDIS_ADDR': Redis server address.
- 'REDIS_PASSWORD': Redis server password.
- 'KAFKA_BROKER': Kafka broker address.
- 'KAFKA_TOPIC': Kafka topic for incoming messages.
- 'PORT': Port for the HTTP server.
