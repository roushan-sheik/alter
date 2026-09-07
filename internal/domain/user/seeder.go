package user

import (
	"log"

	"gorm.io/gorm"
)

// SeedAdmin checks if the admin user exists, and if not, creates one with the provided credentials.
func SeedAdmin(db *gorm.DB, email, password string) {
	if email == "" || password == "" {
		log.Println("Admin email or password not set in config, skipping admin seeding.")
		return
	}

	var admin User
	err := db.Where("email = ?", email).First(&admin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create the admin user
			admin = User{
				Name:               "Admin",
				Email:              email,
				AuthProvider:       AuthProviderEmail,
				Role:               RoleAdmin,
				LanguagePreference: "en",
				Age:                30, // arbitrary
			}
			if err := admin.HashPassword(password); err != nil {
				log.Fatalf("Failed to hash admin password during seeding: %v", err)
			}
			if err := db.Create(&admin).Error; err != nil {
				log.Fatalf("Failed to seed admin user: %v", err)
			}
			log.Println("Successfully seeded admin account from .env credentials.")
		} else {
			log.Fatalf("Failed to query admin user during seeding: %v", err)
		}
	} else {
		// Admin already exists. Make sure the role is correct just in case.
		if admin.Role != RoleAdmin {
			admin.Role = RoleAdmin
			db.Save(&admin)
			log.Println("Updated existing user to ADMIN role.")
		}
		// Notice we intentionally DO NOT update the password here.
		// If the admin changes their password in the UI, we don't want to reset it on every startup.
	}
}
