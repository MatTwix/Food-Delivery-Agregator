# Food Delivery Aggregator

A microservices-based backend system for a food delivery platform, designed to handle orders from multiple restaurants. This project is built with Go and leverages an event-driven architecture using Kafka and gRPC for robust and scalable inter-service communication.

## Core Concepts

* **Microservices Architecture:** Each business domain (users, restaurants, orders, etc.) is encapsulated in its own dedicated service with a private database, ensuring loose coupling and independent scalability.
* **Event-Driven with Saga Pattern:** Long-running business processes, like handling an order, are orchestrated via asynchronous events using Kafka. This creates a resilient system where services react to events rather than making direct, blocking calls.
* **CQRS (Command Query Responsibility Segregation) Principle:** Services maintain their own local copies of data from other services (e.g., `orders-service` keeps a cache of restaurant data) by subscribing to events. This increases fault tolerance.
* **Synchronous Communication for Critical Queries:** For real-time, critical data validation (like fetching menu prices during order creation), services use direct, synchronous gRPC calls.
* **Centralized Authentication:** A dedicated `users-service` handles user management and JWT issuance. An `api-gateway` validates tokens and passes user identity to downstream services.
* **Authorization at the Service Level:** Each service is responsible for its own authorization logic (e.g., checking if a user owns an order or has an 'admin' role).

## Tech Stack

* **Language:** Go
* **Containerization:** Docker & Docker Compose
* **Messaging:** Apache Kafka
* **Inter-service RPC:** gRPC with Protocol Buffers
* **Database:** PostgreSQL (one instance per service)
* **API Gateway Routing:** Chi (`go-chi/chi`)
* **Authentication:** JSON Web Tokens (JWT)
* **Configuration:** Viper
* **Logging:** `slog` (structured JSON logging)

## System Architecture

### Services

| Service                 | Port (HTTP) | Port (gRPC) | Database        | Description                                                                                             |
| ----------------------- | ----------- | ----------- | --------------- | ------------------------------------------------------------------------------------------------------- |
| **API Gateway**         | `3000`      | -           | -               | Single entry point. Handles routing, authentication, and request proxying.                            |
| **Restaurants Service** | `3001`      | `4040`      | `postgres-db`   | Manages restaurant and menu item data. Provides a gRPC endpoint for menu queries.                       |
| **Orders Service**      | `3002`      | -           | `orders-db`     | Manages the entire order lifecycle. Acts as the central orchestrator for the order processing saga.       |
| **Couriers Service**    | `3003`      | -           | `couriers-db`   | Manages couriers, their availability, and assignment to orders.                                         |
| **Users Service**       | `3004`| -           | `users-db`      | Manages user registration, login, password hashing, and JWT generation/refresh.                         |
| **Payments Service**    | `(internal)`| -           | -               | Simulates payment processing. Subscribes to `order.created` events and publishes payment outcomes.        |
| **Notifications Service**| `(internal)`| -           | -               | Subscribes to various system events to simulate sending notifications to users.                          |

---

## Inter-Service Communication (Saga Pattern)

The primary business flow—processing an order—is handled via a chain of Kafka events:

1. **`POST /api/orders/orders`** -> **Orders Service**
    * Validates the order via a gRPC call to `Restaurants Service`.
    * Saves the order with `pending` status.
    * Publishes **`order.created`**.

2. **Payments Service** consumes `order.created`.
    * Simulates payment processing.
    * Publishes **`payment.succeeded`** or **`payment.failed`**.

3. **Orders Service** consumes `payment.succeeded`.
    * Updates order status to `paid`.
    * Publishes **`order.paid`**.

4. **Couriers Service** consumes `order.paid`.
    * Finds an available courier and assigns them.
    * Publishes **`order.courier_assigned`** or **`courier.search.failed`**.

5. **Orders Service** consumes `order.courier_assigned`.
    * Updates order status to `awaiting_pickup` and stores the `courier_id`.

6. **`POST /api/couriers/orders/{id}/delivered`** -> **Couriers Service**
    * Publishes **`order.delivered`**.

7. **Orders Service** consumes `order.delivered`.
    * Updates order status to `delivered`.

8. **Couriers Service** also consumes `order.delivered`.
    * Updates the courier's status to `available`.
    * Publishes **`courier.became_available`**.

9. **Notifications Service** consumes all major events (`order.created`, `payment.succeeded`, etc.) to log simulated notifications.

---

## API Endpoints

All endpoints are accessed through the API Gateway on port `3000`.

### Public Endpoints

* `POST /api/users/register` - Register a new user.
* `POST /api/users/login` - Log in and receive JWTs.
* `POST /api/users/refresh` - Refresh an expired access token.
* `GET /api/restaurants/restaurants` - Get a list of all restaurants.
* `GET /api/restaurants/restaurants/{id}` - Get details for a single restaurant.
* `GET /api/restaurants/menu-items/restaurant/{id}` - Get menu items for a restaurant.

### Protected Endpoints (require `Authorization: Bearer <token>`)

* **Users**
  * `GET /api/users/users` - Get all users (Admin only)
* **Restaurants**
  * `POST /api/restaurants/restaurants` - Create a new restaurant (Admin only).
  * `POST /api/restaurants/menu-items/restaurant/{id}` - Add a menu item (Admin/Owner only).
* **Orders**
  * `POST /api/orders/orders` - Create a new order.
  * `GET /api/orders/orders` - Get a list of your orders.
  * `GET /api/orders/orders/{id}` - Get details for a specific order (Owner or Admin only).
* **Couriers**
  * `POST /api/couriers/orders/{orderId}/picked_up` - Mark an order as picked up by courier.
  * `POST /api/couriers/orders/{orderId}/delivered` - Mark an order as delivered.

## How to Run

1. **Clone the repository.**
2. **Create a `.env` file** in the root directory by copying `.env.example`.
3. **Generate a JWT secret** and add it to the `.env` file:

    ```bash
    openssl rand -base64 32
    ```

4. **Run the system:**

    ```bash
    docker-compose up --build
    ```

5. The API Gateway will be available at `http://localhost:3000`.
