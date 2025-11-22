# URL Shortener Service (shor.ty)

A simple URL shortening service written in Go.
The application exposes an HTTP API for generating short links and resolving them back to the original URL.

---

# 🚀 Quick Start

1. Clone the repository and create `.env` file from provided example:
    ```bash
    make setup
    ```
2. Build and start the app, PostgreSQL, and Redis:

    ```bash
    make up-build
    ```

    Once the stack is up, the API will be available at: `http://localhost:3000`

3. Run database Migrations
    ```bash
    make db-migrate
    ```

---

# 📡 API Overview

## Create a short link
```http request
GET http://localhost:3000/shorten
body: {"long_url": "https://example.com"}
```

## Redirect using a short code
```http request
GET http://localhost:3000/123456
```
The service responds with a redirect to the original URL.