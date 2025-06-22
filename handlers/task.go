package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/kcansari/task-management-api/database"
	"github.com/kcansari/task-management-api/middleware"
	"github.com/kcansari/task-management-api/models"
)

// CreateTaskRequest represents the data needed to create a new task
type CreateTaskRequest struct {
	Title       string             `json:"title"`       // Task title (required)
	Description string             `json:"description"` // Task description (optional)
	Status      models.TaskStatus  `json:"status"`      // Task status (optional, defaults to pending)
}

// UpdateTaskRequest represents the data that can be updated for a task
type UpdateTaskRequest struct {
	Title       *string            `json:"title,omitempty"`       // Pointer allows nil for "not provided"
	Description *string            `json:"description,omitempty"` // Pointer allows nil for "not provided"
	Status      *models.TaskStatus `json:"status,omitempty"`      // Pointer allows nil for "not provided"
}

// TaskResponse represents a task in API responses
type TaskResponse struct {
	ID          uint               `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	Status      models.TaskStatus  `json:"status"`
	UserID      uint               `json:"user_id"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

// PaginatedTaskResponse represents a paginated list of tasks
type PaginatedTaskResponse struct {
	Tasks      []TaskResponse `json:"tasks"`       // The actual task data
	Page       int           `json:"page"`        // Current page number (1-based)
	PageSize   int           `json:"page_size"`   // Number of items per page
	Total      int64         `json:"total"`       // Total number of tasks
	TotalPages int           `json:"total_pages"` // Total number of pages
	HasNext    bool          `json:"has_next"`    // Whether there's a next page
	HasPrev    bool          `json:"has_prev"`    // Whether there's a previous page
}

// GetTasks handles GET /api/tasks - Get all tasks for authenticated user with pagination
func GetTasks(w http.ResponseWriter, r *http.Request) {
	// Set JSON content type
	w.Header().Set("Content-Type", "application/json")

	// Only allow GET method
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Get authenticated user from context (set by middleware)
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		// This should never happen if middleware is working correctly
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found in context"})
		return
	}

	// Parse pagination parameters from query string
	// URL format: /api/tasks?page=2&page_size=10
	query := r.URL.Query()
	
	// Default pagination values
	page := 1
	pageSize := 10 // Default page size
	maxPageSize := 100 // Maximum allowed page size to prevent abuse

	// Parse page parameter
	if pageStr := query.Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse page_size parameter
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
			// Enforce maximum page size to prevent performance issues
			if pageSize > maxPageSize {
				pageSize = maxPageSize
			}
		}
	}

	// Calculate offset for database query
	// OFFSET = (page - 1) * pageSize
	// Example: page 2 with size 10 = offset 10
	offset := (page - 1) * pageSize

	// Get database connection
	db := database.GetDB()
	
	// Count total tasks for this user (needed for pagination metadata)
	var total int64
	if err := db.Model(&models.Task{}).Where("user_id = ?", user.UserID).Count(&total).Error; err != nil {
		log.Printf("Failed to count tasks for user %d: %v", user.UserID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch tasks"})
		return
	}

	// Query tasks with pagination
	// LIMIT controls how many records to return
	// OFFSET controls how many records to skip
	// ORDER BY ensures consistent ordering across pages
	var tasks []models.Task
	if err := db.Where("user_id = ?", user.UserID).
		Order("created_at DESC"). // Most recent first
		Limit(pageSize).
		Offset(offset).
		Find(&tasks).Error; err != nil {
		log.Printf("Failed to fetch tasks for user %d: %v", user.UserID, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to fetch tasks"})
		return
	}

	// Convert models to response format
	taskResponses := make([]TaskResponse, 0)
	for _, task := range tasks {
		taskResponses = append(taskResponses, TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
			UserID:      task.UserID,
			CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	// Calculate pagination metadata
	// Total pages = ceiling(total / pageSize)
	// In Go, integer division truncates, so we add (pageSize-1) to get ceiling effect
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	// Check if there are more pages
	hasNext := page < totalPages
	hasPrev := page > 1

	// Create paginated response
	response := PaginatedTaskResponse{
		Tasks:      taskResponses,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}

	// Return paginated tasks
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTask handles GET /api/tasks/{id} - Get specific task by ID
func GetTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Get authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found in context"})
		return
	}

	// Extract task ID from URL path
	// URL format: /api/tasks/123
	// We need to parse the ID from the path
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Task ID is required"})
		return
	}

	// Convert string ID to integer
	// strconv.ParseUint converts string to unsigned integer
	taskID, err := strconv.ParseUint(path, 10, 32) // base 10, 32-bit uint
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid task ID"})
		return
	}

	// Get database connection
	db := database.GetDB()
	
	// Find task by ID and user ID (for security)
	// This ensures users can only access their own tasks
	var task models.Task
	if err := db.Where("id = ? AND user_id = ?", taskID, user.UserID).First(&task).Error; err != nil {
		// Task not found or doesn't belong to user
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Task not found"})
		return
	}

	// Convert to response format
	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		UserID:      task.UserID,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateTask handles POST /api/tasks - Create a new task
func CreateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Get authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found in context"})
		return
	}

	// Parse request body
	var req CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Validate required fields
	if strings.TrimSpace(req.Title) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Title is required"})
		return
	}

	// Validate status if provided
	if req.Status != "" {
		// Check if status is one of the valid values
		validStatuses := []models.TaskStatus{
			models.TaskStatusPending,
			models.TaskStatusInProgress,
			models.TaskStatusCompleted,
		}
		
		valid := false
		for _, validStatus := range validStatuses {
			if req.Status == validStatus {
				valid = true
				break
			}
		}
		
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid status. Use: pending, in_progress, or completed"})
			return
		}
	} else {
		// Set default status if not provided
		req.Status = models.TaskStatusPending
	}

	// Create new task
	task := models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		UserID:      user.UserID, // Associate task with authenticated user
	}

	// Save to database
	db := database.GetDB()
	if err := db.Create(&task).Error; err != nil {
		log.Printf("Failed to create task: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create task"})
		return
	}

	// Convert to response format
	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		UserID:      task.UserID,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.WriteHeader(http.StatusCreated) // 201 Created
	json.NewEncoder(w).Encode(response)
}

// UpdateTask handles PUT /api/tasks/{id} - Update existing task
func UpdateTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Get authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found in context"})
		return
	}

	// Extract task ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Task ID is required"})
		return
	}

	taskID, err := strconv.ParseUint(path, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid task ID"})
		return
	}

	// Parse request body
	var req UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid JSON"})
		return
	}

	// Find existing task
	db := database.GetDB()
	var task models.Task
	if err := db.Where("id = ? AND user_id = ?", taskID, user.UserID).First(&task).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Task not found"})
		return
	}

	// Update fields if provided (partial update)
	// Using pointers allows us to distinguish between "not provided" and "empty string"
	if req.Title != nil {
		if strings.TrimSpace(*req.Title) == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Title cannot be empty"})
			return
		}
		task.Title = *req.Title
	}

	if req.Description != nil {
		task.Description = *req.Description
	}

	if req.Status != nil {
		// Validate status
		validStatuses := []models.TaskStatus{
			models.TaskStatusPending,
			models.TaskStatusInProgress,
			models.TaskStatusCompleted,
		}
		
		valid := false
		for _, validStatus := range validStatuses {
			if *req.Status == validStatus {
				valid = true
				break
			}
		}
		
		if !valid {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid status. Use: pending, in_progress, or completed"})
			return
		}
		
		task.Status = *req.Status
	}

	// Save updated task
	if err := db.Save(&task).Error; err != nil {
		log.Printf("Failed to update task: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to update task"})
		return
	}

	// Convert to response format
	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		UserID:      task.UserID,
		CreatedAt:   task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteTask handles DELETE /api/tasks/{id} - Delete a task
func DeleteTask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Method not allowed"})
		return
	}

	// Get authenticated user from context
	user, ok := middleware.GetUserFromContext(r)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "User not found in context"})
		return
	}

	// Extract task ID from URL
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Task ID is required"})
		return
	}

	taskID, err := strconv.ParseUint(path, 10, 32)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid task ID"})
		return
	}

	// Find and delete task
	db := database.GetDB()
	var task models.Task
	if err := db.Where("id = ? AND user_id = ?", taskID, user.UserID).First(&task).Error; err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Task not found"})
		return
	}

	// Soft delete the task (GORM sets deleted_at timestamp)
	if err := db.Delete(&task).Error; err != nil {
		log.Printf("Failed to delete task: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to delete task"})
		return
	}

	// Return success with no content
	w.WriteHeader(http.StatusNoContent) // 204 No Content
}