package controllers

import (
	"net/http"
	"strconv"

	"cms/internal/middleware"
	"cms/internal/models"
	"cms/internal/repositories"
	"cms/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userRepo     *repositories.UserRepository
	employeeRepo *repositories.EmployeeRepository
}

func NewUserController(userRepo *repositories.UserRepository, employeeRepo *repositories.EmployeeRepository) *UserController {
	return &UserController{
		userRepo:     userRepo,
		employeeRepo: employeeRepo,
	}
}

func (ctl *UserController) ListUsers(c *gin.Context) {
	role := c.GetString("role")
	companyID, ok := middleware.CompanyIDFromContext(c)

	// Fallback: if companyID not in token, get from user record
	if !ok {
		userID, userOk := middleware.UserIDFromContext(c)
		if userOk {
			user, err := ctl.userRepo.GetUserByID(userID)
			if err == nil && user != nil {
				companyID = user.CompanyID
				ok = true
			}
		}
	}

	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		c.Abort()
		return
	}

	var users []interface{}

	// Super admin sees all users, admin sees only their company users
	if role == "super_admin" {
		allUsers, err := ctl.userRepo.GetAllUsers()
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		users = make([]interface{}, len(allUsers))
		for i, u := range allUsers {
			users[i] = u
		}
	} else {
		companyUsers, err := ctl.userRepo.ListUsersByCompany(companyID)
		if err != nil {
			utils.Error(c, http.StatusInternalServerError, err.Error())
			return
		}
		users = make([]interface{}, len(companyUsers))
		for i, u := range companyUsers {
			users[i] = u
		}
	}

	utils.Success(c, http.StatusOK, users)
}

func (ctl *UserController) CreateUser(c *gin.Context) {
	role := c.GetString("role")
	companyID, ok := middleware.CompanyIDFromContext(c)

	// Fallback: if companyID not in token, get from user record
	if !ok {
		userID, userOk := middleware.UserIDFromContext(c)
		if userOk {
			user, err := ctl.userRepo.GetUserByID(userID)
			if err == nil && user != nil {
				companyID = user.CompanyID
				ok = true
			}
		}
	}

	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		c.Abort()
		return
	}

	// Only admins and super_admins can create users
	if role != "admin" && role != "super_admin" {
		utils.Error(c, http.StatusForbidden, "only admins can create users")
		return
	}

	// Get admin ID (the one creating this user)
	adminIDAny, ok := c.Get(middleware.ContextUserIDKey)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "admin id not found")
		c.Abort()
		return
	}
	adminID, ok := adminIDAny.(uint)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "invalid admin id")
		c.Abort()
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if email already exists
	_, err := ctl.userRepo.GetUserByEmail(req.Email)
	if err == nil {
		utils.Error(c, http.StatusConflict, "email already exists")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "failed to hash password")
		return
	}

	// Create user
	user := &models.User{
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		Role:      req.Role,
		Status:    req.Status,
		CompanyID: companyID,
		CreatedBy: adminID,
	}

	if err := ctl.userRepo.CreateUser(user); err != nil {
		handleError(c, err)
		return
	}

	// Auto-create Employee record so user appears in employee lists
	employee := &models.Employee{
		Name:      user.Name,
		Position:  "Employee",
		Role:      user.Role,
		Salary:    0,
		CompanyID: companyID,
		UserID:    &user.ID,
	}

	if err := ctl.employeeRepo.Create(employee); err != nil {
		// If employee creation fails, still return the user (don't block)
		// The user is created successfully, just not appearing in employee lists yet
	}

	utils.Success(c, http.StatusCreated, user)
}

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=120"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=admin employee super_admin manager salesman developer staff"`
	Status   string `json:"status" binding:"required,oneof=active inactive pending"`
}

type UpdateUserRequest struct {
	Name   string `json:"name" binding:"required,min=2,max=120"`
	Email  string `json:"email" binding:"required,email"`
	Role   string `json:"role" binding:"required,oneof=admin employee super_admin manager salesman developer staff"`
	Status string `json:"status" binding:"required,oneof=active inactive pending"`
}

func (ctl *UserController) UpdateUser(c *gin.Context) {
	role := c.GetString("role")
	companyID, ok := middleware.CompanyIDFromContext(c)
	if !ok {
		utils.Error(c, http.StatusUnauthorized, "company not found in token")
		c.Abort()
		return
	}

	// Only admins and super_admins can update users
	if role != "admin" && role != "super_admin" {
		utils.Error(c, http.StatusForbidden, "only admins can update users")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Get the user to update
	user, err := ctl.userRepo.GetUserByID(uint(id))
	if err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}

	// Admins can only update users in their company
	if role == "admin" && user.CompanyID != companyID {
		utils.Error(c, http.StatusForbidden, "cannot update users from other companies")
		return
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Role = req.Role
	user.Status = req.Status

	if err := ctl.userRepo.UpdateUser(user); err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, user)
}

func (ctl *UserController) DeleteUser(c *gin.Context) {
	role := c.GetString("role")

	// Only super_admin can delete users
	if role != "super_admin" {
		utils.Error(c, http.StatusForbidden, "only super_admin can delete users")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "invalid user id")
		return
	}

	// Get the user to delete
	user, err := ctl.userRepo.GetUserByID(uint(id))
	if err != nil {
		utils.Error(c, http.StatusNotFound, "user not found")
		return
	}

	if err := ctl.userRepo.DeleteUser(user.ID); err != nil {
		handleError(c, err)
		return
	}

	utils.Success(c, http.StatusOK, gin.H{
		"message": "user deleted successfully",
	})
}
