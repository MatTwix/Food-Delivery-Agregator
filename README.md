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

#### User Authentication

* **`POST /api/users/register`** - Register a new user
  * **Request Body:**

    ```json
    {
      "email": "user@example.com",
      "password": "password123"
    }
    ```

  * **Response:** User object with `id`, `email`, `role`, `created_at`, `updated_at`

* **`POST /api/users/login`** - Log in and receive JWTs
  * **Request Body:**

    ```json
    {
      "email": "user@example.com",
      "password": "password123"
    }
    ```

  * **Response:**

    ```json
    {
      "access_token": "jwt_access_token_here",
      "refresh_token": "jwt_refresh_token_here"
    }
    ```

* **`POST /api/users/refresh`** - Refresh an expired access token
  * **Request Body:**

    ```json
    {
      "refresh_token": "jwt_refresh_token_here"
    }
    ```

  * **Response:**

    ```json
    {
      "access_token": "new_jwt_access_token_here",
      "refresh_token": "new_jwt_refresh_token_here"
    }
    ```

#### Restaurants & Menu

* **`GET /api/restaurants/restaurants`** - Get a list of all restaurants
  * **Response:** Array of restaurant objects with `id`, `owner_id`, `name`, `address`, `phone_number`, `created_at`, `updated_at`

* **`GET /api/restaurants/restaurants/{id}`** - Get details for a single restaurant
  * **Response:** Single restaurant object

* **`GET /api/restaurants/menu-items`** - Get menu items for all restaurants
  * **Response:** Array of menu item objects with `id`, `restaurant_id`, `name`, `description`, `price`, `created_at`, `updated_at`

* **`GET /api/restaurants/menu-items/restaurant/{id}`** - Get menu items for a restaurant
  * **Response:** Array of menu item objects for specified restaurant

#### Health Checks

* `GET /api/restaurants/health` - Check restaurants service status
* `GET /api/orders/health` - Check orders service status
* `GET /api/couriers/health` - Check couriers service status
* `GET /api/users/health` - Check users service status

### Protected Endpoints (require `Authorization: Bearer <token>`)

#### Users Management

* **`GET /api/users/users`** - Get all users (Admin only)
  * **Response:** Array of user objects

* **`POST /api/users/users/{id}/role`** - Assign user a role (Admin only)
  * **Request Body:**

    ```json
    {
      "role": "courier",
      "name": "John Doe"
    }
    ```

  * **Response:**

    ```json
    {
      "user_id": "user_uuid",
      "prev_role": "user",
      "new_role": "courier",
      "name": "John Doe"
    }
    ```

* **`DELETE /api/users/users/{id}`** - Delete user (Admin only)
  * **Response:** Success message

#### Restaurant Management

* **`POST /api/restaurants/restaurants`** - Create a new restaurant (Admin/Manager only)
  * **Request Body:**

    ```json
    {
      "owner_id": "owner_uuid",
      "name": "Pizza Palace",
      "address": "123 Main St",
      "phone_number": "+1234567890"
    }
    ```

  * **Response:** Created restaurant object with generated `id`

* **`PUT /api/restaurants/restaurants/{id}`** - Update restaurant (Admin/Manager/Owner only)
  * **Request Body:** Same as creation request
  * **Response:** Updated restaurant object

* **`DELETE /api/restaurants/restaurants/{id}`** - Delete restaurant (Admin/Manager/Owner only)
  * **Response:** Success message

#### Menu Management

* **`POST /api/restaurants/menu-items/restaurant/{id}`** - Add a menu item (Admin/Manager/Restaurant owner only)
  * **Request Body:**

    ```json
    {
      "name": "Margherita Pizza",
      "description": "Classic tomato and mozzarella",
      "price": 12.99
    }
    ```

  * **Response:** Created menu item object with generated `id`

* **`PUT /api/restaurants/menu-items/{id}`** - Update menu item (Admin/Manager/Restaurant owner only)
  * **Request Body:** Same as creation request
  * **Response:** Updated menu item object

* **`DELETE /api/restaurants/menu-items/{id}`** - Delete menu item (Admin/Manager/Restaurant owner only)
  * **Response:** Success message

#### Order Management

* **`POST /api/orders/orders`** - Create a new order
  * **Request Body:**

    ```json
    {
      "restaurant_id": "restaurant_uuid",
      "items": [
        {
          "menu_item_id": "menu_item_uuid",
          "quantity": 2
        },
        {
          "menu_item_id": "another_menu_item_uuid",
          "quantity": 1
        }
      ]
    }
    ```

  * **Response:** Created order object with `id`, `restaurant_id`, `user_id`, `total_price`, `status`, `courier_id`, `items[]`, `created_at`, `updated_at`

* **`GET /api/orders/orders`** - Get all orders (Admin/Manager only)
  * **Response:** Array of order objects

* **`GET /api/orders/orders/{id}`** - Get specific order (Admin/Manager/Owner only)
  * **Response:** Single order object

#### Courier Management

* **`GET /api/couriers/couriers`** - Get all couriers (Admin only)
  * **Response:** Array of courier objects with `id`, `name`, `status`, `created_at`, `updated_at`

* **`GET /api/couriers/couriers/available`** - Get available courier (Admin only)
  * **Response:** Single available courier object or 404 if none available

* **`PUT /api/couriers/couriers/{id}`** - Update courier (Admin only)
  * **Request Body:**

    ```json
    {
      "name": "John Courier",
      "status": "available"
    }
    ```

  * **Response:** Updated courier object

#### Order Delivery

* **`POST /api/couriers/orders/{orderId}/picked_up`** - Mark order as picked up (Admin/Courier only)
  * **Response:** Success message

* **`POST /api/couriers/orders/{orderId}/delivered`** - Mark order as delivered (Admin/Courier only)
  * **Response:** Success message

---

## Kafka Events & Topics

The system uses Apache Kafka for asynchronous inter-service communication. Here are the topics and event structures:

### Core Business Events

#### Order Events

* **Topic:** `order.created`
  * **Producer:** Orders Service
  * **Consumers:** Payments Service, Notifications Service
  * **Event Structure:**

    ```json
    {
      "id": "order_uuid",
      "restaurant_id": "restaurant_uuid",
      "user_id": "user_uuid",
      "total_price": 25.99,
      "status": "pending",
      "courier_id": null,
      "items": [
        {
          "id": "order_item_uuid",
          "order_id": "order_uuid",
          "menu_item_id": "menu_item_uuid",
          "quantity": 2,
          "price": 12.99
        }
      ],
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
    ```

* **Topic:** `order.paid`
  * **Producer:** Orders Service
  * **Consumers:** Couriers Service, Notifications Service
  * **Event Structure:** Same as `order.created` but with `status: "paid"`

* **Topic:** `order.picked_up`
  * **Producer:** Couriers Service
  * **Consumers:** Notifications Service
  * **Event Structure:**

    ```json
    {
      "courier_id": "courier_uuid",
      "order_id": "order_uuid",
      "user_id": "user_uuid"
    }
    ```

* **Topic:** `order.delivered`
  * **Producer:** Couriers Service
  * **Consumers:** Orders Service, Couriers Service, Notifications Service
  * **Event Structure:**

    ```json
    {
      "courier_id": "courier_uuid",
      "order_id": "order_uuid",
      "user_id": "user_uuid"
    }
    ```

#### Payment Events

* **Topic:** `payment.succeeded`
  * **Producer:** Payments Service
  * **Consumers:** Orders Service, Notifications Service
  * **Event Structure:**

    ```json
    {
      "id": "order_uuid",
      "user_id": "user_uuid",
      "total_price": 25.99
    }
    ```

* **Topic:** `payment.failed`
  * **Producer:** Payments Service
  * **Consumers:** Orders Service, Notifications Service
  * **Event Structure:** Same as `payment.succeeded`

#### Courier Events

* **Topic:** `courier.assigned`
  * **Producer:** Couriers Service
  * **Consumers:** Orders Service
  * **Event Structure:**

    ```json
    {
      "courier_id": "courier_uuid"
    }
    ```

* **Topic:** `courier.search.failed`
  * **Producer:** Couriers Service
  * **Consumers:** Orders Service
  * **Event Structure:**

    ```json
    {
      "id": "order_uuid"
    }
    ```

### Restaurant Management Events

* **Topic:** `restaurant.created`
  * **Producer:** Restaurants Service
  * **Consumers:** Orders Service (for local cache)
  * **Event Structure:**

    ```json
    {
      "id": "restaurant_uuid",
      "owner_id": "owner_uuid",
      "name": "Pizza Palace",
      "address": "123 Main St",
      "phone_number": "+1234567890",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
    ```

* **Topic:** `restaurant.updated`
  * **Producer:** Restaurants Service
  * **Consumers:** Orders Service (for local cache)
  * **Event Structure:** Same as `restaurant.created`

* **Topic:** `restaurant.deleted`
  * **Producer:** Restaurants Service
  * **Consumers:** Orders Service (for local cache)
  * **Event Structure:**

    ```json
    {
      "id": "restaurant_uuid"
    }
    ```

### User Management Events

* **Topic:** `users.role_assigned`
  * **Producer:** Users Service
  * **Consumers:** Couriers Service
  * **Event Structure:**

    ```json
    {
      "user_id": "user_uuid",
      "prev_role": "user",
      "new_role": "courier",
      "name": "John Doe"
    }
    ```

### Notification Events

* **Topic:** `notification.created`
  * **Producer:** Notifications Service
  * **Consumers:** External notification systems (webhooks, email services, etc.)
  * **Event Structure:**

    ```json
    {
      "user_id": "user_uuid",
      "order_id": "order_uuid",
      "message": "Order has been created."
    }
    ```

### Event Processing Flow

The **Notifications Service** acts as a centralized event processor that:

1. **Consumes** multiple event types:
   * `order.created`
   * `payment.succeeded`
   * `payment.failed`
   * `order.picked_up`
   * `order.delivered`

2. **Transforms** each consumed event into a standardized `NotificationEvent`

3. **Publishes** to `notification.created` topic for external consumption

4. **Maps** events to user-friendly messages:
   * `order.created` → "Order has been created."
   * `payment.succeeded` → "Payment was successful."
   * `payment.failed` → "Payment failed."
   * `order.picked_up` → "Order picked up by courier."
   * `order.delivered` → "Order delivered."

---

## Port Configuration

The following ports are used by the system when running with Docker Compose:

### Application Services

| Service | Port | Protocol | Purpose | Access |
|---------|------|----------|---------|--------|
| **API Gateway** | `3000` | HTTP | Main application entry point | Public |
| **Restaurants Service** | `3001` | HTTP | Restaurant management API | Internal |
| **Restaurants Service** | `4040` | gRPC | Menu item queries | Internal |
| **Orders Service** | `3002` | HTTP | Order management API | Internal |
| **Couriers Service** | `3003` | HTTP | Courier management API | Internal |
| **Users Service** | `3004` | HTTP | User authentication & management | Internal |

### Infrastructure Services

| Service | Port | Protocol | Purpose | Access |
|---------|------|----------|---------|--------|
| **Kafka** | `9092` | TCP | Internal Kafka broker | Internal |
| **Kafka** | `29092` | TCP | External Kafka access | External |
| **Kafdrop** | `19000` | HTTP | Kafka Web UI | Development |
| **Zookeeper** | `2181` | TCP | Kafka coordination | Internal |

### Database Services

| Service | Port | Protocol | Database | Access |
|---------|------|----------|----------|--------|
| **postgres-db** | `5432` | TCP | Restaurants data | Development |
| **orders-db** | `5433` | TCP | Orders data | Development |
| **couriers-db** | `5434` | TCP | Couriers data | Development |
| **users-db** | `5435` | TCP | Users data | Development |

### Port Usage Summary

**Total Ports Used:** 11 ports

**Production Access:**

* `3000` - Main API Gateway (the only port you need to expose in production)

**Development Access:**

* `5432-5435` - PostgreSQL databases (for direct database access during development)
* `19000` - Kafdrop Web UI (for monitoring Kafka topics and messages)
* `29092` - Kafka external access (for external Kafka clients)

**Internal Only:**

* `3001-3004` - Microservices (accessed only through API Gateway in production)
* `4040` - gRPC endpoint (internal service communication)
* `9092` - Kafka broker (internal communication)
* `2181` - Zookeeper (Kafka coordination)

### Network Security Notes

* **Production:** Only port `3000` (API Gateway) should be publicly accessible
* **Development:** All ports can be accessible on localhost for debugging
* **Internal Services:** Ports `3001-3004` and `4040` are for inter-service communication
* **Database Ports:** `5432-5435` should only be accessible from application services in production

---

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
