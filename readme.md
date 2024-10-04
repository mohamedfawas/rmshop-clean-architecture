# Real Madrid Shop API

## Overview

This project is a robust e-commerce API for a Sports team merchandise shop, built using Go and following clean architecture principles. It provides a set of endpoints for managing users, products, categories, and orders in a secure and efficient manner.

## Features

- Clean Architecture design
- User authentication and authorization
- Product management (CRUD operations)
- Category and subcategory management
- Shopping cart functionality
- Order processing
- Wishlist management
- User profile and address management
- Admin panel for store management
- Coupon system
- Payment integration (Razorpay)
- Inventory management
- Sales reporting and analytics
- Return and refund processing
- Order cancellation and refund processing
- Wallet system to manage refund
- Email verification for user registration

## Technology Stack

- Go (Golang)
- PostgreSQL
- JWT for authentication
- Gorilla Mux for routing
- SMTP for email sending
- Cloudinary for image storage
- Razorpay for payment processing

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

4. Set up your PostgreSQL database and update the configuration in `config.yaml`.

5. Run database migrations:
   ```
   go run cmd/migrate/main.go
   ```

6. Build and run the application:
   ```
   go run cmd/api/main.go
   ```

## API Endpoints

For a complete list of endpoints and their descriptions, please refer to the [API documentation](https://documenter.getpostman.com/view/17830038/2sAXxLDaYW).

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
