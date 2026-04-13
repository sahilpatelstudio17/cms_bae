package seed

import (
	"log"
	"time"

	"cms/internal/models"
	"cms/internal/utils"

	"gorm.io/gorm"
)

func SeedDatabase(db *gorm.DB) error {
	log.Println("🌱 Starting database seeding...")

	// Hash password (same for all admins: "password123")
	adminPwd, err := utils.HashPassword("password123")
	if err != nil {
		return err
	}

	// Define 4 companies with admin data
	companiesData := []struct {
		companyName string
		adminName   string
		adminEmail  string
	}{
		{
			companyName: "Company One",
			adminName:   "Admin One",
			adminEmail:  "company1@gmail.com",
		},
		{
			companyName: "Company Two",
			adminName:   "Admin Two",
			adminEmail:  "company2@gmail.com",
		},
		{
			companyName: "Company Three",
			adminName:   "Admin Three",
			adminEmail:  "company3@gmail.com",
		},
		{
			companyName: "Company Four",
			adminName:   "Admin Four",
			adminEmail:  "company4@gmail.com",
		},
	}

	var company *models.Company

	// Create 4 companies with admins
	for _, compData := range companiesData {
		// Create company
		comp := &models.Company{
			Name:  compData.companyName,
			Email: "info@" + compData.companyName + ".com",
		}
		if err := db.Create(comp).Error; err != nil {
			log.Printf("❌ Error creating company: %v\n", err)
			return err
		}
		log.Printf("✅ Company created: %s (ID: %d)\n", comp.Name, comp.ID)

		// Store first company for additional seeding later
		if company == nil {
			company = comp
		}

		// Create subscription for company
		subscription := &models.Subscription{
			CompanyID: comp.ID,
			Plan:      "pro",
			Status:    "active",
		}
		if err := db.Create(subscription).Error; err != nil {
			log.Printf("❌ Error creating subscription: %v\n", err)
			return err
		}

		// Create admin user for company
		adminUser := &models.User{
			Name:      compData.adminName,
			Email:     compData.adminEmail,
			Password:  adminPwd,
			Role:      "admin",
			CompanyID: comp.ID,
		}
		if err := db.Create(adminUser).Error; err != nil {
			log.Printf("❌ Error creating admin user %s: %v\n", adminUser.Email, err)
			return err
		}
		log.Printf("✅ Admin created: %s (%s) - Email: %s - Password: password123\n", adminUser.Name, adminUser.Role, adminUser.Email)
	}

	// Continue with additional seeding for first company only
	if company == nil {
		log.Println("❌ No company was created")
		return err
	}

	// Hash passwords for other users
	superAdminPwd, err := utils.HashPassword("SuperAdmin@123")
	if err != nil {
		return err
	}
	employeePwd, err := utils.HashPassword("Employee@123")
	if err != nil {
		return err
	}

	// Create additional users for the first company (for testing)
	users := []models.User{
		{
			Name:      "System SuperAdmin",
			Email:     "superadmin@company.com",
			Password:  superAdminPwd,
			Role:      "super_admin",
			CompanyID: company.ID,
		},
		{
			Name:      "Alice Employee",
			Email:     "alice@company.com",
			Password:  employeePwd,
			Role:      "employee",
			CompanyID: company.ID,
		},
		{
			Name:      "Bob Employee",
			Email:     "bob@company.com",
			Password:  employeePwd,
			Role:      "employee",
			CompanyID: company.ID,
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			log.Printf("❌ Error creating user %s: %v\n", user.Email, err)
			return err
		}
		log.Printf("✅ User created: %s (%s) - Email: %s\n", user.Name, user.Role, user.Email)
	}

	// Create employees
	employees := []models.Employee{
		{
			Name:      "Alice Employee",
			Position:  "Software Engineer",
			Salary:    75000,
			CompanyID: company.ID,
		},
		{
			Name:      "Bob Employee",
			Position:  "Product Manager",
			Salary:    85000,
			CompanyID: company.ID,
		},
		{
			Name:      "Charlie Davis",
			Position:  "Designer",
			Salary:    70000,
			CompanyID: company.ID,
		},
		{
			Name:      "Diana Wilson",
			Position:  "QA Engineer",
			Salary:    72000,
			CompanyID: company.ID,
		},
	}

	for _, emp := range employees {
		if err := db.Create(&emp).Error; err != nil {
			log.Printf("❌ Error creating employee %s: %v\n", emp.Name, err)
			return err
		}
		log.Printf("✅ Employee created: %s (%s)\n", emp.Name, emp.Position)
	}

	// Fetch employees for task assignment
	var emps []models.Employee
	if err := db.Where("company_id = ?", company.ID).Find(&emps).Error; err != nil {
		return err
	}

	// Create tasks
	if len(emps) > 0 {
		tasks := []models.Task{
			{
				Title:       "Design homepage",
				Description: "Create a modern design for the homepage",
				Status:      "in_progress",
				AssignedTo:  emps[0].ID,
				CompanyID:   company.ID,
			},
			{
				Title:       "Fix login bug",
				Description: "User reports issues with password reset",
				Status:      "pending",
				AssignedTo:  emps[1].ID,
				CompanyID:   company.ID,
			},
			{
				Title:       "Write API documentation",
				Description: "Document all REST endpoints",
				Status:      "completed",
				AssignedTo:  emps[0].ID,
				CompanyID:   company.ID,
			},
			{
				Title:       "Performance optimization",
				Description: "Optimize database queries",
				Status:      "in_progress",
				AssignedTo:  emps[2].ID,
				CompanyID:   company.ID,
			},
		}

		for _, task := range tasks {
			if err := db.Create(&task).Error; err != nil {
				log.Printf("❌ Error creating task %s: %v\n", task.Title, err)
				return err
			}
			log.Printf("✅ Task created: %s\n", task.Title)
		}
	}

	// Create attendance records
	today := time.Now()
	for i := 0; i < 7; i++ {
		for _, emp := range emps {
			attendance := models.Attendance{
				EmployeeID: emp.ID,
				CompanyID:  company.ID,
				Date:       today.AddDate(0, 0, -i),
			}

			if err := db.Create(&attendance).Error; err != nil {
				log.Printf("❌ Error creating attendance: %v\n", err)
				continue
			}
		}
	}
	log.Println("✅ Attendance records created")

	// Create pending users for approval workflow demonstration
	pendingUserPwd, err := utils.HashPassword("PendingUser@123")
	if err != nil {
		return err
	}

	pendingUser := &models.User{
		Name:      "Charlie Pending",
		Email:     "charlie.pending@example.com",
		Password:  pendingUserPwd,
		Role:      "employee",
		Status:    "pending",
		CompanyID: company.ID,
	}

	if err := db.Create(pendingUser).Error; err != nil {
		log.Printf("❌ Error creating pending user: %v\n", err)
		return err
	}
	log.Printf("✅ Pending user created: %s - Email: %s (Status: pending)\n", pendingUser.Name, pendingUser.Email)

	// Create approval request for the pending user
	approval := &models.ApprovalRequest{
		RequestType: "user",
		UserID:      pendingUser.ID,
		CompanyID:   company.ID,
		Status:      "pending",
		Message:     "Awaiting admin approval",
	}

	if err := db.Create(approval).Error; err != nil {
		log.Printf("❌ Error creating approval request: %v\n", err)
		return err
	}
	log.Printf("✅ Approval request created for user: %s\n", pendingUser.Name)

	// Create admin approval request for testing super admin approvals page
	adminPwdHash, err := utils.HashPassword("AdminPassword@123")
	if err != nil {
		return err
	}

	adminApproval := &models.ApprovalRequest{
		RequestType:    "admin",
		UserID:         0, // Admin hasn't been created yet
		CompanyID:      company.ID,
		CompanyName:    company.Name,
		Status:         "pending",
		Message:        "New admin account creation request",
		AdminName:      "David Admin",
		AdminEmail:     "david.admin@example.com",
		RequestedEmail: "david.admin@example.com",
		PasswordHash:   adminPwdHash, // Store hashed password for when admin approves
	}

	if err := db.Create(adminApproval).Error; err != nil {
		log.Printf("❌ Error creating admin approval request: %v\n", err)
		return err
	}
	log.Printf("✅ Admin approval request created for: %s\n", adminApproval.AdminEmail)

	log.Println("\n🎉 Database seeding completed successfully!")
	log.Println("\n📝 Test Credentials:")
	log.Println("   Super Admin:")
	log.Println("   - Email: superadmin@company.com")
	log.Println("   - Password: SuperAdmin@123")
	log.Println("   - Role: Can create/manage Admins")
	log.Println("\n   Admin:")
	log.Println("   - Email: company1@gmail.com")
	log.Println("   - Password: password123")
	log.Println("   - Role: Can create/manage Employees and Approvals")
	log.Println("\n   Employee (Active):")
	log.Println("   - Email: alice@company.com")
	log.Println("   - Password: Employee@123")
	log.Println("   - Role: Can view tasks and mark attendance")
	log.Println("\n   Employee (Pending Approval - Test Approval Workflow):")
	log.Println("   - Email: charlie.pending@example.com")
	log.Println("   - Password: PendingUser@123")
	log.Println("   - Status: PENDING (Cannot login until admin approves)")
	log.Println("   - Admin can approve/reject at /approvals page")
	log.Println("\n   Admin Approval Request (Test Super Admin Approvals):")
	log.Println("   - Name: David Admin")
	log.Println("   - Email: david.admin@example.com")
	log.Println("   - Password: AdminPassword@123")
	log.Println("   - Status: PENDING (Super admin can approve to create admin user)")
	log.Println("   - Super admin can approve/reject at /approvals page")

	return nil
}

func CleanDatabase(db *gorm.DB) error {
	log.Println("🗑️  Cleaning database...")

	// Delete in order of dependencies (respecting foreign keys)
	if err := db.Exec("DELETE FROM approval_requests").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM attendances").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM tasks").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM expenses").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM sales").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM employees").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM users").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM subscriptions").Error; err != nil {
		return err
	}
	if err := db.Exec("DELETE FROM companies").Error; err != nil {
		return err
	}

	log.Println("✅ Database cleaned")
	return nil
}
