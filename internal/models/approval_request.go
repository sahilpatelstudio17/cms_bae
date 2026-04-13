package models

import "time"

type ApprovalRequest struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	RequestType    string    `json:"request_type" gorm:"size:30;not null"` // "user", "role_assignment", "admin", or "employee"
	UserID         uint      `json:"user_id" gorm:"index"`                 // For user approvals (0 = not set for company signups)
	User           *User     `json:"user" gorm:"-"`                        // Manual load, no FK constraint
	EmployeeID     uint      `json:"employee_id" gorm:"index"`             // For role assignment requests
	Employee       *Employee `json:"employee" gorm:"-"`                    // Manual load, no FK constraint
	RequestedRole  string    `json:"requested_role" gorm:"size:50"`        // Role being requested
	RequestedEmail string    `json:"requested_email" gorm:"size:120"`      // Email for the user to be created
	RequestedBy    uint      `json:"requested_by" gorm:"index"`            // Admin who created request
	CompanyID      uint      `json:"company_id" gorm:"index;not null"`
	ApprovedBy     uint      `json:"approved_by" gorm:"index"`              // Admin who approved
	Status         string    `json:"status" gorm:"size:30;default:pending"` // pending, approved, rejected
	Message        string    `json:"message" gorm:"type:text"`              // Reason for approval/rejection
	// Additional fields for company signup approval
	CompanyName  string `json:"company_name" gorm:"size:120"` // For company signup requests
	AdminName    string `json:"admin_name" gorm:"size:120"`   // For company signup requests
	AdminEmail   string `json:"admin_email" gorm:"size:120"`  // For company signup requests
	PasswordHash string `json:"-" gorm:"size:255"`            // Hashed password (not exposed in JSON)
	Password     string `json:"password" gorm:"-"`            // Password (not stored, used only during creation)
	// Additional fields for employee approval
	RequestedData   string    `json:"requested_data" gorm:"size:255"`  // For employee: position, or other data
	Salary          float64   `json:"salary" gorm:"default:0"`         // For employee approval: salary
	RequestedStatus string    `json:"requested_status" gorm:"size:30"` // User's requested status (for user requests)
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
