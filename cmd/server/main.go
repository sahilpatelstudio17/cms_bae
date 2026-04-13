package main

import (
	"log"

	"cms/config"
	"cms/internal/controllers"
	"cms/internal/middleware"
	"cms/internal/repositories"
	"cms/internal/services"
	"cms/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	db := config.NewDatabase(cfg.DatabaseURL, cfg.Environment == "production")

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.CORS())

	authRepo := repositories.NewAuthRepository(db)
	companyRepo := repositories.NewCompanyRepository(db)
	employeeRepo := repositories.NewEmployeeRepository(db)
	taskRepo := repositories.NewTaskRepository(db)
	attendanceRepo := repositories.NewAttendanceRepository(db)
	userRepo := repositories.NewUserRepository(db)
	approvalRepo := repositories.NewApprovalRequestRepository(db)
	expenseRepo := repositories.NewExpenseRepository(db)
	saleRepo := repositories.NewSaleRepository(db)

	authService := services.NewAuthService(authRepo, approvalRepo, cfg.JWTSecret, cfg.JWTExpiresHours)
	companyService := services.NewCompanyService(companyRepo)
	employeeService := services.NewEmployeeService(employeeRepo)
	taskService := services.NewTaskService(taskRepo, employeeRepo)
	attendanceService := services.NewAttendanceService(attendanceRepo, employeeRepo, userRepo)
	userManagementService := services.NewUserManagementService(userRepo, employeeRepo, cfg.JWTSecret, cfg.JWTExpiresHours)
	approvalService := services.NewApprovalService(userRepo, approvalRepo, companyRepo, authRepo, employeeRepo, cfg.JWTSecret, cfg.JWTExpiresHours)
	expenseService := services.NewExpenseService(expenseRepo, userRepo, employeeRepo)
	salesService := services.NewSalesService(saleRepo, userRepo, employeeRepo)
	dashboardService := services.NewDashboardService(employeeRepo, taskRepo, attendanceRepo)
	bulkImportService := services.NewBulkImportService(userRepo, employeeRepo)
	userImportService := services.NewUserImportService(userRepo, employeeRepo)
	roleAssignmentService := services.NewRoleAssignmentService(approvalRepo, employeeRepo, userRepo, companyRepo)

	authController := controllers.NewAuthController(authService, approvalService)
	companyController := controllers.NewCompanyController(companyService)
	employeeController := controllers.NewEmployeeController(employeeService, userRepo, userManagementService)
	taskController := controllers.NewTaskController(taskService)
	attendanceController := controllers.NewAttendanceController(attendanceService)
	adminController := controllers.NewAdminController(userManagementService)
	approvalController := controllers.NewApprovalController(approvalService)
	expenseController := controllers.NewExpenseController(expenseService)
	salesController := controllers.NewSalesController(salesService)
	dashboardController := controllers.NewDashboardController(dashboardService)
	bulkImportController := controllers.NewBulkImportController(bulkImportService, userImportService)
	userController := controllers.NewUserController(userRepo, employeeRepo)
	roleAssignmentController := controllers.NewRoleAssignmentController(roleAssignmentService)

	routes.RegisterRoutes(router, routes.Dependencies{
		AuthController:           authController,
		CompanyController:        companyController,
		EmployeeController:       employeeController,
		TaskController:           taskController,
		AttendanceController:     attendanceController,
		AdminController:          adminController,
		ApprovalController:       approvalController,
		ExpenseController:        expenseController,
		SalesController:          salesController,
		DashboardController:      dashboardController,
		BulkImportController:     bulkImportController,
		UserController:           userController,
		RoleAssignmentController: roleAssignmentController,
		JWTSecret:                cfg.JWTSecret,
	})

	log.Printf("server running on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}
