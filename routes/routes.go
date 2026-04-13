package routes

import (
	"cms/internal/controllers"
	"cms/internal/middleware"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	AuthController           *controllers.AuthController
	CompanyController        *controllers.CompanyController
	EmployeeController       *controllers.EmployeeController
	TaskController           *controllers.TaskController
	AttendanceController     *controllers.AttendanceController
	AdminController          *controllers.AdminController
	ApprovalController       *controllers.ApprovalController
	ExpenseController        *controllers.ExpenseController
	SalesController          *controllers.SalesController
	DashboardController      *controllers.DashboardController
	BulkImportController     *controllers.BulkImportController
	UserController           *controllers.UserController
	RoleAssignmentController *controllers.RoleAssignmentController
	JWTSecret                string
}

func RegisterRoutes(router *gin.Engine, deps Dependencies) {
	api := router.Group("/api")

	auth := api.Group("/auth")
	auth.POST("/register", deps.AuthController.Register)
	auth.POST("/register-with-approval", deps.AuthController.RegisterWithApproval)
	auth.POST("/request-admin-approval", deps.AuthController.RequestAdminApproval)
	auth.POST("/login", deps.AuthController.Login)

	protected := api.Group("")
	protected.Use(middleware.JWTAuth(deps.JWTSecret))

	// Auth protected routes
	auth.GET("/me", middleware.JWTAuth(deps.JWTSecret), deps.AuthController.GetMe)
	auth.PUT("/profile", middleware.JWTAuth(deps.JWTSecret), deps.AuthController.UpdateProfile)
	auth.POST("/change-password", middleware.JWTAuth(deps.JWTSecret), deps.AuthController.ChangePassword)

	// Admin management routes (Super Admin only)
	admins := protected.Group("/admins")
	admins.GET("", middleware.RequireRoles("super_admin"), deps.AdminController.ListAdmins)
	admins.POST("", middleware.RequireRoles("super_admin"), deps.AdminController.CreateAdmin)
	admins.DELETE("/:id", middleware.RequireRoles("super_admin"), deps.AdminController.DeleteAdmin)

	company := protected.Group("/company")
	company.GET("", deps.CompanyController.Get)
	company.PUT("", middleware.RequireRoles("admin"), deps.CompanyController.Update)

	dashboard := protected.Group("/dashboard")
	dashboard.GET("/stats", deps.DashboardController.GetStats)

	employees := protected.Group("/employees")
	employees.GET("", deps.EmployeeController.List)
	employees.POST("", middleware.RequireRoles("admin", "manager"), deps.EmployeeController.Create)
	employees.POST("/bulk-import", middleware.RequireRoles("admin"), deps.BulkImportController.ImportEmployees)
	employees.POST("/with-user", middleware.RequireRoles("admin", "manager"), deps.EmployeeController.CreateWithUser)
	employees.PUT("/:id", middleware.RequireRoles("admin", "manager"), deps.EmployeeController.Update)
	employees.DELETE("/:id", middleware.RequireRoles("admin", "manager"), deps.EmployeeController.Delete)

	tasks := protected.Group("/tasks")
	tasks.GET("", deps.TaskController.List)
	tasks.POST("", middleware.RequireRoles("admin", "manager"), deps.TaskController.Create)
	tasks.PUT("/:id", deps.TaskController.Update)
	tasks.DELETE("/:id", middleware.RequireRoles("admin", "manager"), deps.TaskController.Delete)

	approvals := protected.Group("/approvals")
	// Specific routes FIRST (before parameterized routes)
	// Super admin routes
	approvals.GET("/admin/pending", middleware.RequireRoles("super_admin"), deps.ApprovalController.GetPendingAdminRequests)
	approvals.POST("/admin/:id/approve", middleware.RequireRoles("super_admin"), deps.ApprovalController.ApproveAdminRequest)
	approvals.POST("/admin/:id/reject", middleware.RequireRoles("super_admin"), deps.ApprovalController.RejectAdminRequest)
	// Company signup routes (specific route before general :id routes)
	approvals.GET("/company/pending", middleware.RequireRoles("admin"), deps.ApprovalController.GetPendingCompanySignups)
	approvals.POST("/company/:id/approve", middleware.RequireRoles("admin"), deps.ApprovalController.ApproveCompanySignup)
	// Employee request routes
	approvals.POST("/employee/request", middleware.RequireRoles("admin"), deps.ApprovalController.RequestEmployeeApproval)
	approvals.GET("/employee/pending", middleware.RequireRoles("admin"), deps.ApprovalController.GetPendingEmployeeRequests)
	approvals.POST("/employee/:id/approve", middleware.RequireRoles("admin"), deps.ApprovalController.ApproveEmployeeRequest)
	approvals.POST("/employee/:id/reject", middleware.RequireRoles("admin"), deps.ApprovalController.RejectEmployeeRequest)
	// User request routes
	approvals.POST("/user/request", middleware.RequireRoles("admin"), deps.ApprovalController.RequestUserApproval)
	approvals.GET("/user/pending", middleware.RequireRoles("admin"), deps.ApprovalController.GetPendingUserApprovals)
	approvals.POST("/user/:id/approve", middleware.RequireRoles("admin"), deps.ApprovalController.ApproveUserRequest)
	approvals.POST("/user/:id/reject", middleware.RequireRoles("admin"), deps.ApprovalController.RejectUserRequest)

	// General approval routes (all authenticated users can view with role-based filtering)
	approvals.GET("", deps.ApprovalController.ListPendingApprovals)
	approvals.POST("/:id/approve", middleware.RequireRoles("admin"), deps.ApprovalController.ApproveUser)
	approvals.POST("/:id/reject", middleware.RequireRoles("admin"), deps.ApprovalController.RejectUser)

	expenses := protected.Group("/expenses")
	expenses.GET("", deps.ExpenseController.ListExpenses)
	expenses.POST("", deps.ExpenseController.CreateExpense)
	expenses.GET("/pending", middleware.RequireRoles("admin", "manager"), deps.ExpenseController.ListPendingApprovals)
	expenses.POST("/:id/approve", middleware.RequireRoles("admin", "manager"), deps.ExpenseController.ApproveExpense)
	expenses.POST("/:id/reject", middleware.RequireRoles("admin", "manager"), deps.ExpenseController.RejectExpense)

	sales := protected.Group("/sales")
	sales.GET("/pending", middleware.RequireRoles("admin", "manager"), deps.SalesController.ListPendingApprovals)
	sales.GET("", deps.SalesController.ListSales)
	sales.POST("", deps.SalesController.CreateSale)
	sales.PUT("/:id", deps.SalesController.UpdateSale)
	sales.DELETE("/:id", deps.SalesController.DeleteSale)
	sales.POST("/:id/approve", middleware.RequireRoles("admin", "manager"), deps.SalesController.ApproveSale)
	sales.POST("/:id/reject", middleware.RequireRoles("admin", "manager"), deps.SalesController.RejectSale)

	users := protected.Group("/users")
	users.GET("", deps.UserController.ListUsers)
	users.POST("", middleware.RequireRoles("admin", "super_admin"), deps.UserController.CreateUser)
	users.POST("/bulk-import", middleware.RequireRoles("admin", "super_admin"), deps.BulkImportController.ImportUsers)
	users.PUT("/:id", deps.UserController.UpdateUser)
	users.DELETE("/:id", deps.UserController.DeleteUser)

	// Role assignment approval requests
	roleAssignments := protected.Group("/role-assignments")
	roleAssignments.GET("/pending", middleware.RequireRoles("super_admin", "admin"), deps.RoleAssignmentController.GetPendingRoleAssignments)
	roleAssignments.POST("/request", middleware.RequireRoles("admin"), deps.RoleAssignmentController.RequestRoleAssignment)
	roleAssignments.POST("/:id/approve", middleware.RequireRoles("super_admin", "admin"), deps.RoleAssignmentController.ApproveRoleAssignment)
	roleAssignments.POST("/:id/reject", middleware.RequireRoles("super_admin", "admin"), deps.RoleAssignmentController.RejectRoleAssignment)
}
