package dto

type RegisterRequest struct {
	Name                    string `json:"name" example:"John Doe"     validate:"required,min=2,max=100"`
	Email                   string `json:"email" example:"user@example.com"    validate:"required,email"`
	Password                string `json:"password" example:"Secret123!" validate:"required,min=8"`
	AgreeTermsAndConditions bool   `json:"agreeTermsAndConditions" example:"true" validate:"required"`
	// Available languages: en (English), fr (French), es (Spanish), pt (Portuguese), ht (Haitian Creole)
	LanguagePreference      string `json:"language_preference" example:"en" enums:"en,fr,es,pt,ht" validate:"omitempty,oneof=en fr es pt ht"`
	Age                     int    `json:"age" example:"25" validate:"required,min=0,max=120"`
}

type LoginRequest struct {
	Email              string `json:"email" example:"user@example.com"    validate:"required,email"`
	Password           string `json:"password" example:"Secret123!" validate:"required"`
	// Available languages: en (English), fr (French), es (Spanish), pt (Portuguese), ht (Haitian Creole)
	LanguagePreference string `json:"language_preference" example:"en" enums:"en,fr,es,pt,ht" validate:"omitempty,oneof=en fr es pt ht"`
}

type AdminLoginRequest struct {
	Email              string `json:"email" example:"admin@altar.com"    validate:"required,email"`
	Password           string `json:"password" example:"Admin1234" validate:"required"`
	// Available languages: en (English), fr (French), es (Spanish), pt (Portuguese), ht (Haitian Creole)
	LanguagePreference string `json:"language_preference" example:"en" enums:"en,fr,es,pt,ht" validate:"omitempty,oneof=en fr es pt ht"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIUzI1NiIsIn..."`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" example:"user@example.com" validate:"required,email"`
}

type ResendOTPRequest struct {
	Email string `json:"email" example:"user@example.com" validate:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" example:"user@example.com" validate:"required,email"`
	OTP   string `json:"otp" example:"12345"          validate:"required,len=5"`
}

type ResetPasswordRequest struct {
	ResetToken  string `json:"reset_token" example:"eyJhb..." validate:"required"`
	NewPassword string `json:"new_password" example:"NewSecret123!" validate:"required,min=8"`
}

type UpdateProfileRequest struct {
	Name               *string `json:"name" example:"John Doe"                omitempty:"true" validate:"omitempty,min=2,max=100"`
	Location           *string `json:"location" example:"New York, USA"            omitempty:"true"`
	// Available themes: LIGHT, DARK
	ThemePreference    *string `json:"theme_preference" example:"DARK" enums:"LIGHT,DARK" omitempty:"true" validate:"omitempty,oneof=LIGHT DARK"`
	// Available languages: en (English), fr (French), es (Spanish), pt (Portuguese), ht (Haitian Creole)
	LanguagePreference *string `json:"language_preference" example:"en" enums:"en,fr,es,pt,ht" omitempty:"true" validate:"omitempty,oneof=en fr es pt ht"`
	Age                *int    `json:"age" example:"25" omitempty:"true" validate:"omitempty,min=0,max=120"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" example:"Secret123!"     validate:"required"`
	NewPassword     string `json:"new_password" example:"NewSecret123!"     validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" example:"NewSecret123!" validate:"required,eqfield=NewPassword"`
}

type RegisterDeviceRequest struct {
	Token    string `json:"token" example:"fcm-token-123"    validate:"required"`
	Platform string `json:"platform" example:"IOS" validate:"required,oneof=IOS ANDROID"`
}

type SocialLoginRequest struct {
	IDToken  string `json:"id_token" example:"eyJhbGciOi..." validate:"required"`
	Provider string `json:"provider" example:"google" validate:"required,oneof=google apple"`
	Name     string `json:"name" example:"John Doe"`
	Email    string `json:"email" example:"user@example.com" validate:"required,email"`
}
