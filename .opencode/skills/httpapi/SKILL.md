
---
name: httpapi
description: Design and implement maintainable Go HTTP APIs with separated server configuration, endpoints, business logic, and OpenAPI documentation.
---

# HTTP API Architecture Skill

## Purpose

Ensure HTTP APIs follow a consistent, maintainable architecture with clear separation of concerns.

## 1. Architecture and separation of concerns

Always separate the following responsibilities:

### Server configuration
Responsible for HTTP server setup and lifecycle.

Includes:
- HTTP server configuration.
- Router initialization.
- Middleware registration.
- Graceful shutdown.
- Read, write and idle timeouts.
- Dependency wiring.
- Server startup and shutdown.

Do not put endpoint business logic in server configuration.

### Endpoints (HTTP handlers)

Endpoints are responsible only for HTTP concerns.

Every endpoint should:
1. Validate and parse the incoming request.
2. Convert the request into the appropriate service input.
3. Call the corresponding service.
4. Handle and translate errors into appropriate HTTP responses.
5. Return the HTTP response.

Handlers must not contain business logic or directly access repositories or databases.

Keep handlers thin and focused.

### Endpoint logic (Services)

Business logic must live in dedicated services, independently of HTTP.

Services are responsible for:
- Business rules and validation.
- Orchestrating domain operations.
- Calling repositories and external dependencies.
- Returning domain results and errors.

Services must not depend on HTTP request or response types.

## 2. OpenAPI documentation

Always provide OpenAPI documentation for every endpoint.

Document:
- HTTP method and path.
- Endpoint description and operation ID.
- Path, query and header parameters.
- Request body and schemas.
- Response schemas and status codes.
- Authentication and authorization requirements.
- Possible error responses.

Keep documentation synchronized with the actual implementation.

Avoid undocumented endpoints and undocumented response formats.

## 3. Middleware

Use middleware for cross-cutting HTTP concerns, such as:
- Authentication and authorization: confirm with me if is necesary or not.
- Request ID and correlation IDs.
- Structured request logging.
- Panic recovery.
- CORS: confirm with me if is necesary or not.
- Rate limiting: confirm with me if is necesary or not.
- Request Body Size Limit
- Request timeouts.

Do not duplicate middleware responsibilities inside individual handlers.

## 4. Error handling

- Use consistent error mapping between application errors and HTTP status codes.
- Do not expose internal errors, stack traces or sensitive information.
- Keep domain errors independent of HTTP status codes.
- Return consistent error response structures.
- Always log unexpected internal errors with sufficient context.

## 5. Testing

Add tests for:
- Handler request validation.
- Successful responses.
- Invalid requests.
- Service errors and HTTP error mapping.
- Middleware behavior.
- OpenAPI specification consistency.

Prefer table-driven tests where appropriate.

## 6. Decision-making rules

- Follow the existing project architecture and conventions.
- Do not introduce unnecessary abstractions.
- For super-simple applications, a simpler structure may be appropriate.
- If the application is super simple and you believe a different architecture or separation is preferable, ASK ME BEFORE MAKING THAT DECISION.
- Do not silently combine server configuration, handlers and business logic just to reduce the number of files.
- Ask before introducing new external dependencies.
- Use go-playground/validator to validate the inputs.

## 7. Completion checklist

Before considering an HTTP API implementation complete, verify:

- [ ] Server configuration is separated from endpoint handlers.
- [ ] Middleware is registered centrally.
- [ ] Graceful shutdown is implemented.
- [ ] Handlers only validate, call services and handle errors.
- [ ] Business logic lives in services.
- [ ] Every endpoint has OpenAPI documentation.
- [ ] Tests cover handlers and relevant service behavior.
- [ ] OpenAPI documentation matches the implementation.
