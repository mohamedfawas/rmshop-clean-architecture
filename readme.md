# Real Madrid Shop API

## Overview

This project is a robust e-commerce API for a Real Madrid merchandise shop, built using Go and following clean architecture principles. It provides a set of endpoints for managing users, products, categories, and orders in a secure and efficient manner.

## Features

- Clean Architecture design
- User authentication and authorization
- Product management (CRUD operations)
- Category and subcategory management
- Order processing
- Admin panel functionality
- Email verification for user registration
- JWT-based authentication
- PostgreSQL database integration

## Technology Stack

- Go (Golang)
- PostgreSQL
- JWT for authentication
- Gorilla Mux for routing
- SMTP for email sending

## Project Structure

The project follows a clean architecture pattern:

```
rmshop-clean-architecture/
│
├── cmd/
│   ├── api/
│   │   └── main.go                 # Main application entry point
│   └── seedadmin/
│       └── main.go                 # Admin seeding command
│
├── internal/
│   ├── config/
│   ├── delivery/
│   │   └── http/
│   ├── domain/
│   ├── repository/
│   ├── usecase/
│   └── server/
│
├── migrations/
├── pkg/
│   ├── auth/
│   ├── database/
│   ├── email/
│   └── validator/
│
├── scripts/
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

## Setup and Installation

1. Clone the repository:
   ```
   git clone https://github.com/yourusername/rmshop-clean-architecture.git
   ```

2. Navigate to the project directory:
   ```
   cd rmshop-clean-architecture
   ```

3. Install dependencies:
   ```
   go mod tidy
   ```

4. Set up your PostgreSQL database and update the `.env` file with your database credentials.

5. Run database migrations:
   ```
   ./scripts/run_migrations.sh
   ```

6. Start the server:
   ```
   go run cmd/api/main.go
   ```

## API Endpoints

The API provides the following main endpoints:

- User Management:
  - POST /user/signup
  - POST /user/verify-otp
  - POST /user/resend-otp
  - POST /user/login
  - POST /user/logout

- Admin Management:
  - POST /admin/login
  - POST /admin/logout

- Category Management:
  - POST /admin/categories
  - GET /admin/categories
  - GET /admin/categories/{categoryId}
  - PUT /admin/categories/{categoryId}
  - DELETE /admin/categories/{categoryId}

- Subcategory Management:
  - POST /admin/categories/{categoryId}/subcategories
  - GET /admin/categories/{categoryId}/subcategories
  - GET /admin/categories/{categoryId}/subcategories/{subcategoryId}
  - PUT /admin/categories/{categoryId}/subcategories/{subcategoryId}
  - DELETE /admin/categories/{categoryId}/subcategories/{subcategoryId}

- Product Management:
  - POST /admin/products
  - GET /admin/products
  - GET /admin/products/{productId}
  - PUT /admin/products/{productId}
  - DELETE /admin/products/{productId}

For a complete list of endpoints and their descriptions, please refer to the API documentation.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
