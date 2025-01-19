package dto

type RegisterUserDTO struct {
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Password    string `json:"password" validate:"required"`
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name"`
	DisplayName string `json:"display_name"`
	Country     string `json:"country"`
	PublicKey   []byte `json:"public_key" validate:"required"`
}

type LoginUserDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	Salt     string `json:"salt" validate:"required"`
}

type LoginClientDTO struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type LoginPrecheckDTO struct {
	Username string `json:"username" validate:"required"`
}

type UpdateDisplayNameDTO struct {
	DisplayName string `json:"display_name" validate:"required"`
}

type RegisterClientDTO struct {
	Name        string `json:"name" validate:"required"`
	RedirectURI string `json:"redirect_uri" validate:"required"`
	BackendURI  string `json:"backend_uri" validate:"required"`
	Username    string `json:"username" validate:"required,min=3,max=50"`
	Password    string `json:"password" validate:"required"`
}

type CheckBackendURIDTO struct {
	BackendURI string `json:"backend_uri" validate:"required"`
}

type VerifyEmailDTO struct {
	Email string `json:"email" validate:"required,email"`
}

type CheckEmailVerificationCodeDTO struct {
	Email string `json:"email" validate:"required,email"`
	Code  string `json:"code" validate:"required"`
}

type ResetPasswordDTO struct {
	Username    string `json:"username" validate:"required"`
	Signature   []byte `json:"signature" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type RegisterUserPrecheckDTO struct {
	Username string `json:"username" validate:"required"`
}