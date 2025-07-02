# Basic GO authentication implementation

Implementation of JWT authentication in Golang as a first time practice.

## Endpoints

| Endpoint        | Method | Description             |
| --------------- | ------ | ----------------------- |
| /register       | POST   | User registration       |
| /login          | POST   | User login              |
| /refresh_tokens | GET    | Refresh JWT tokens      |
| /logout         | GET    | User logout             |
| /user           | GET    | Basic user access point |

_Swagger_ is used to generated documentation: `/swagger`

## Entities

- _users_ (id, username, password, created_at)
- _refresh_tokens_ (id, token, userId, keyPairId, userAgent, agentIp, issued_at, expires_at)

## Running

_Environment variables_:

1.  Database

    - DB_HOST
    - DB_PORT
    - DB_USER
    - DB_PASSWORD
    - DB_NAME
    - DB_SSLMODE

2.  Authentication

    - JWT_SECRET
    - JWT_ACCESS_EXPIRATION
    - JWT_REFRESH_EXPIRATION
    - JWT_REFRESH_LENGTH

3.  App

    - SERVER_PORT

The code is run via `docker-compose`

```bash
docker-compose -f docker-compose.yaml up -d
```

Occupied ports:

- 8080 - app
- 5432 - database server
- 9000 - webhook server
