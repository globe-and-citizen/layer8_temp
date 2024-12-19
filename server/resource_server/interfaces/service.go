package interfaces

import (
	"globe-and-citizen/layer8/server/resource_server/dto"
	"globe-and-citizen/layer8/server/resource_server/models"
)

type IService interface {
	RegisterUser(req dto.RegisterUserDTO) error
	LoginPreCheckUser(req dto.LoginPrecheckDTO) (models.LoginPrecheckResponseOutput, error)
	LoginPreCheckClient(req dto.LoginPrecheckDTO) (models.LoginPrecheckResponseOutput, error)
	LoginUser(req dto.LoginUserDTO) (models.LoginUserResponseOutput, error)
	LoginClient(req dto.LoginClientDTO) (models.LoginUserResponseOutput, error)
	ProfileUser(userID uint) (models.ProfileResponseOutput, error)
	ProfileClient(userID string) (models.ClientResponseOutput, error)
	FindUser(userID uint) (models.User, error)
	VerifyEmail(userID uint, userEmail string) error
	CheckEmailVerificationCode(userID uint, code string) error
	GenerateZkProofOfEmailVerification(
		user models.User, request dto.CheckEmailVerificationCodeDTO,
	) ([]byte, uint, error)
	SaveProofOfEmailVerification(userID uint, verificationCode string, zkProof []byte, zkKeyPairId uint) error
	UpdateDisplayName(userID uint, req dto.UpdateDisplayNameDTO) error
	RegisterClient(req dto.RegisterClientDTO) error
	GetClientData(clientName string) (models.ClientResponseOutput, error)
	GetClientDataByBackendURL(backendURL string) (models.ClientResponseOutput, error)
	CheckBackendURI(backendURL string) (bool, error)
	GetUserForUsername(username string) (models.User, error)
	ValidateSignature(message string, signature []byte, publicKey []byte) error
	UpdateUserPassword(username string, newPassword string, salt string) error
	RegisterUserPrecheck(req dto.RegisterUserPrecheckDTO, iterCount string) (models.RegisterUserPrecheckResponseOutput, error)
}
