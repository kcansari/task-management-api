# Task Management API Documentation

## Overview

A RESTful API for task management built with Go, featuring user authentication, CRUD operations for tasks, and comprehensive security measures.

**Base URL**: `http://localhost:8080`

**Authentication**: JWT Bearer tokens

## Table of Contents

1. [Authentication](#authentication)
2. [Tasks](#tasks)
3. [Error Handling](#error-handling)
4. [Pagination](#pagination)
5. [Examples](#examples)

## Authentication

All task-related endpoints require authentication via JWT tokens. Include the token in the `Authorization` header:

```
Authorization: Bearer <your-jwt-token>
```

### Register User

Create a new user account.

**Endpoint**: `POST /api/auth/register`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response** (201 Created):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "created_at": "2025-06-22T17:30:00Z",
    "updated_at": "2025-06-22T17:30:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid JSON or missing required fields
- `409 Conflict`: Email already exists

### Login User

Authenticate existing user and receive JWT token.

**Endpoint**: `POST /api/auth/login`

**Request Body**:
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response** (200 OK):
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "created_at": "2025-06-22T17:30:00Z",
    "updated_at": "2025-06-22T17:30:00Z"
  }
}
```

**Error Responses**:
- `400 Bad Request`: Invalid JSON or missing required fields
- `401 Unauthorized`: Invalid email or password

## Tasks

All task endpoints require authentication. Users can only access their own tasks.

### Get Tasks (with Pagination)

Retrieve tasks for the authenticated user with pagination support.

**Endpoint**: `GET /api/tasks`

**Query Parameters**:
- `page` (optional): Page number (default: 1)
- `page_size` (optional): Items per page (default: 10, max: 100)

**Example**: `GET /api/tasks?page=2&page_size=5`

**Headers**:
```
Authorization: Bearer <your-jwt-token>
```

**Response** (200 OK):
```json
{
  "tasks": [
    {
      "id": 1,
      "title": "Complete project documentation",
      "description": "Write comprehensive API documentation",
      "status": "in_progress",
      "user_id": 1,
      "created_at": "2025-06-22T17:30:00+03:00",
      "updated_at": "2025-06-22T17:45:00+03:00"
    }
  ],
  "page": 1,
  "page_size": 10,
  "total": 15,
  "total_pages": 2,
  "has_next": true,
  "has_prev": false
}
```

### Get Single Task

Retrieve a specific task by ID.

**Endpoint**: `GET /api/tasks/{id}`

**Headers**:
```
Authorization: Bearer <your-jwt-token>
```

**Response** (200 OK):
```json
{
  "id": 1,
  "title": "Complete project documentation",
  "description": "Write comprehensive API documentation",
  "status": "in_progress",
  "user_id": 1,
  "created_at": "2025-06-22T17:30:00+03:00",
  "updated_at": "2025-06-22T17:45:00+03:00"
}
```

**Error Responses**:
- `404 Not Found`: Task doesn't exist or doesn't belong to user
- `400 Bad Request`: Invalid task ID format

### Create Task

Create a new task for the authenticated user.

**Endpoint**: `POST /api/tasks`

**Headers**:
```
Authorization: Bearer <your-jwt-token>
Content-Type: application/json
```

**Request Body**:
```json
{
  "title": "New task title",
  "description": "Task description (optional)",
  "status": "pending"
}
```

**Task Status Values**:
- `pending` (default)
- `in_progress`
- `completed`

**Response** (201 Created):
```json
{
  "id": 2,
  "title": "New task title",
  "description": "Task description (optional)",
  "status": "pending",
  "user_id": 1,
  "created_at": "2025-06-22T18:00:00+03:00",
  "updated_at": "2025-06-22T18:00:00+03:00"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid JSON, missing title, or invalid status

### Update Task

Update an existing task (partial updates supported).

**Endpoint**: `PUT /api/tasks/{id}`

**Headers**:
```
Authorization: Bearer <your-jwt-token>
Content-Type: application/json
```

**Request Body** (all fields optional):
```json
{
  "title": "Updated title",
  "description": "Updated description",
  "status": "completed"
}
```

**Response** (200 OK):
```json
{
  "id": 1,
  "title": "Updated title",
  "description": "Updated description",
  "status": "completed",
  "user_id": 1,
  "created_at": "2025-06-22T17:30:00+03:00",
  "updated_at": "2025-06-22T18:15:00+03:00"
}
```

**Error Responses**:
- `404 Not Found`: Task doesn't exist or doesn't belong to user
- `400 Bad Request`: Invalid JSON, empty title, or invalid status

### Delete Task

Delete a task (soft delete - task is marked as deleted but retained in database).

**Endpoint**: `DELETE /api/tasks/{id}`

**Headers**:
```
Authorization: Bearer <your-jwt-token>
```

**Response** (204 No Content): Empty response body

**Error Responses**:
- `404 Not Found`: Task doesn't exist or doesn't belong to user
- `400 Bad Request`: Invalid task ID format

## Error Handling

All endpoints return consistent error responses:

```json
{
  "error": "Human-readable error message"
}
```

### Common HTTP Status Codes

- `200 OK`: Successful GET/PUT request
- `201 Created`: Successful POST request
- `204 No Content`: Successful DELETE request
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Missing/invalid authentication
- `404 Not Found`: Resource not found
- `405 Method Not Allowed`: HTTP method not supported
- `409 Conflict`: Resource conflict (e.g., duplicate email)
- `500 Internal Server Error`: Server error

### Authentication Errors

- Missing Authorization header: `"Authorization header required"`
- Invalid header format: `"Invalid authorization header format"`
- Wrong scheme: `"Invalid authorization scheme. Use Bearer"`
- Invalid/expired token: `"Invalid or expired token"`

## Pagination

The `GET /api/tasks` endpoint supports pagination to handle large datasets efficiently.

### Pagination Parameters

- `page`: Page number (1-based, default: 1)
- `page_size`: Items per page (default: 10, maximum: 100)

### Pagination Response Fields

- `tasks`: Array of task objects for current page
- `page`: Current page number
- `page_size`: Items per page
- `total`: Total number of tasks
- `total_pages`: Total number of pages
- `has_next`: Boolean indicating if there's a next page
- `has_prev`: Boolean indicating if there's a previous page

### Example Pagination Usage

```bash
# Get first page (default)
GET /api/tasks

# Get second page with 5 items per page
GET /api/tasks?page=2&page_size=5

# Get all items on one page (max 100)
GET /api/tasks?page_size=100
```

## Examples

### Complete Workflow Example

```bash
# 1. Register a new user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email": "demo@example.com", "password": "demopass123"}'

# Response includes token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# 2. Create a task
curl -X POST http://localhost:8080/api/tasks \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"title": "Learn Go", "description": "Complete Go tutorial", "status": "in_progress"}'

# 3. Get all tasks
curl -X GET http://localhost:8080/api/tasks \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 4. Update task status
curl -X PUT http://localhost:8080/api/tasks/1 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"status": "completed"}'

# 5. Delete task
curl -X DELETE http://localhost:8080/api/tasks/1 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### JavaScript/Fetch Example

```javascript
// Login and get token
const loginResponse = await fetch('http://localhost:8080/api/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    email: 'demo@example.com',
    password: 'demopass123'
  })
});

const { token } = await loginResponse.json();

// Create task
const createResponse = await fetch('http://localhost:8080/api/tasks', {
  method: 'POST',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    title: 'Learn JavaScript',
    description: 'Complete JavaScript course',
    status: 'pending'
  })
});

// Get tasks with pagination
const tasksResponse = await fetch('http://localhost:8080/api/tasks?page=1&page_size=10', {
  headers: {
    'Authorization': `Bearer ${token}`,
  }
});

const { tasks, total, has_next } = await tasksResponse.json();
```

## Security Features

- **Password Hashing**: Uses bcrypt with proper salt generation
- **JWT Tokens**: 24-hour expiration, signed with HMAC-SHA256
- **Authorization**: Users can only access their own tasks
- **Input Validation**: Comprehensive validation for all endpoints
- **SQL Injection Protection**: GORM provides parameterized queries
- **Rate Limiting**: Page size limited to prevent abuse

## Development Notes

- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT with custom claims (user_id, email)
- **Soft Deletes**: Deleted tasks are marked but not removed
- **Timestamps**: All resources include created_at and updated_at
- **Ordering**: Tasks ordered by creation date (newest first)
- **Environment**: Configurable via .env file