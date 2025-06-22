# Task Management API

A simple REST API for task management built with Go (Golang). This project is designed as a learning exercise to understand Go fundamentals through practical implementation.

## 🚀 Features

- User registration and authentication
- JWT-based authorization
- Task CRUD operations (Create, Read, Update, Delete)
- User-specific task management
- PostgreSQL database integration
- RESTful API design

## 🛠️ Tech Stack

- **Language**: Go (Golang)
- **Web Framework**: Gin
- **Database**: PostgreSQL
- **ORM**: GORM
- **Authentication**: JWT tokens
- **Password Hashing**: bcrypt

## 📋 Prerequisites

Before running this project, make sure you have:

- Go 1.19+ installed
- PostgreSQL database (local or hosted)
- Git for version control
- Postman or curl for API testing

## 🔧 Installation

1. Clone the repository:
```bash
git clone https://github.com/kcansari/task-management-api.git
cd task-management-api
```

2. Install dependencies:
```bash
go mod tidy
```

3. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your database credentials
```

4. Run the application:
```bash
go run main.go
```

The API will be available at `http://localhost:8080`

## 📚 API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user

### Tasks (Protected Routes)
- `GET /api/tasks` - Get all tasks for authenticated user
- `GET /api/tasks/:id` - Get specific task
- `POST /api/tasks` - Create new task
- `PUT /api/tasks/:id` - Update task
- `DELETE /api/tasks/:id` - Delete task

### Users (Protected Routes)
- `GET /api/users/profile` - Get current user profile
- `PUT /api/users/profile` - Update user profile

## 📝 Example Usage

### Register a new user
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "securepassword"
  }'
```

### Create a task (requires authentication token)
```bash
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{
    "title": "Complete project",
    "description": "Finish the task management API",
    "status": "pending"
  }'
```

## 🗂️ Project Structure

```
task-management-api/
├── main.go                 # Application entry point
├── go.mod                  # Go modules file
├── go.sum                  # Dependency lock file
├── .env                    # Environment variables
├── README.md              # This file
│
├── config/                 # Configuration management
│   └── config.go          
│
├── models/                 # Data models
│   ├── user.go            
│   └── task.go            
│
├── database/              # Database connection
│   └── database.go        
│
├── handlers/              # HTTP handlers
│   ├── auth.go           
│   ├── user.go           
│   └── task.go           
│
├── middleware/            # Middleware functions
│   ├── auth.go           
│   └── cors.go           
│
├── utils/                 # Utility functions
│   ├── jwt.go            
│   └── password.go       
│
└── routes/                # Route definitions
    └── routes.go         
```

## 🧪 Testing

Run tests with:
```bash
go test ./...
```

## 📖 Learning Goals

This project teaches:
- Go syntax and idioms
- Struct-based programming
- Interface design patterns
- Error handling patterns
- JWT authentication
- Database operations with GORM
- RESTful API design
- Middleware implementation

## 🚧 Development Status

- ✅ Project setup and structure
- ⏳ Database models and migration
- ⏳ Authentication system
- ⏳ Task CRUD operations
- ⏳ Testing and documentation

## 🤝 Contributing

This is a learning project, but feel free to suggest improvements or report issues.

## 📄 License

This project is licensed under the MIT License.

## 📞 Contact

Created by [kcansari](https://github.com/kcansari) - feel free to contact me!

---

*This project is part of a Go learning journey. Each commit represents a step in understanding Go fundamentals through practical implementation.*