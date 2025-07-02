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


## Running

The code is run via `docker-compose`

```bash
docker-compose -f docker-compose.yaml up -d
```

Occupied ports:

- 8080 - app
- 5432 - database server
- 9000 - webhook server

