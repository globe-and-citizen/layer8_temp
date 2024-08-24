package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"globe-and-citizen/layer8/server/resource_server/db"
	"globe-and-citizen/layer8/server/resource_server/dto"
	"globe-and-citizen/layer8/server/resource_server/interfaces"
	"globe-and-citizen/layer8/server/resource_server/models"
	"globe-and-citizen/layer8/server/resource_server/repository"
	"globe-and-citizen/layer8/server/resource_server/utils"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/public/welcome.html")
}
func LoginUserPage(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/user_portal/login.html")
}
func RegisterUserPage(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/user_portal/register.html")
}
func ClientProfilePage(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/client_portal/profile.html")
}
func UserHandler(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/user_portal/profile.html")
}
func ClientHandler(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/client_portal/register.html")
}
func LoginClientPage(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/client_portal/login.html")
}

func InputYourEmailPage(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/user_portal/email/input-your-email.html")
}

func InputVerificationCodePage(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/user_portal/email/input-verification-code.html")
}

func PasswordResetRequestPage(w http.ResponseWriter, r *http.Request) {
	ServeFileHandler(w, r, "assets-v1/templates/src/pages/user_portal/password_reset/reset-request.html")
}

func ServeFileHandlerWithArgs(
	w http.ResponseWriter,
	r *http.Request,
	filePath string,
	args map[string]interface{},
) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Println(w, http.StatusText(http.StatusMethodNotAllowed))
		return
	}

	args["ProxyURL"] = os.Getenv("PROXY_URL")

	utils.ParseHTML(w, http.StatusOK, filePath, args)
}

func ServeFileHandler(w http.ResponseWriter, r *http.Request, filePath string) {
	ServeFileHandlerWithArgs(w, r, filePath, map[string]interface{}{})
}

func LoginClientHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	request, err := utils.DecodeJsonFromRequest[dto.LoginClientDTO](w, r.Body)
	if err != nil {
		return
	}

	tokenResp, err := newService.LoginClient(request)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to get client profile", err)
		return
	}

	if err := json.NewEncoder(w).Encode(tokenResp); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func ClientProfileHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodGet) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)
	tokenString := r.Header.Get("Authorization")
	tokenString = tokenString[7:]
	clientClaims, err := utils.ValidateClientToken(tokenString)
	if err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "Authentication error: invalid token", err)
		return
	}

	profileResp, err := newService.ProfileClient(clientClaims.UserName)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to get user profile, user not found", err)
		return
	}

	if err := json.NewEncoder(w).Encode(profileResp); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	request, err := utils.DecodeJsonFromRequest[dto.RegisterUserDTO](w, r.Body)
	if err != nil {
		return
	}

	err = newService.RegisterUser(request)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to register user", err)
		return
	}

	res := utils.BuildResponseWithNoBody(w, http.StatusCreated, "User registered successfully")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func RegisterClientHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	request, err := utils.DecodeJsonFromRequest[dto.RegisterClientDTO](w, r.Body)
	if err != nil {
		return
	}

	err = newService.RegisterClient(request)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to register client", err)
		return
	}

	res := utils.BuildResponseWithNoBody(w, http.StatusCreated, "Client registered successfully")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

// LoginPrecheckHandler handles login precheck requests and get the salt of the user from the database using the username from the request URL
func LoginPrecheckHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	request, err := utils.DecodeJsonFromRequest[dto.LoginPrecheckDTO](w, r.Body)
	if err != nil {
		return
	}

	loginPrecheckResp, err := newService.LoginPreCheckUser(request)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to get client profile", err)
		return
	}

	if err := json.NewEncoder(w).Encode(loginPrecheckResp); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	request, err := utils.DecodeJsonFromRequest[dto.LoginUserDTO](w, r.Body)
	if err != nil {
		return
	}

	tokenResp, err := newService.LoginUser(request)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to get client profile", err)
		return
	}

	if err := json.NewEncoder(w).Encode(tokenResp); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodGet) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)
	tokenString := r.Header.Get("Authorization")
	tokenString = tokenString[7:] // Remove the "Bearer " prefix
	userID, err := utils.ValidateToken(tokenString)
	if err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "Authentication error: invalid token", err)
		return
	}

	profileResp, err := newService.ProfileUser(userID)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to get user profile", err)
		return
	}

	if err := json.NewEncoder(w).Encode(profileResp); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func GetClientData(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodGet) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)
	clientName := r.Header.Get("Name")

	clientModel, err := newService.GetClientData(clientName)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to get client profile", err)
		return
	}

	if err := json.NewEncoder(w).Encode(clientModel); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)
	tokenString := r.Header.Get("Authorization")
	tokenString = tokenString[7:] // Remove the "Bearer " prefix
	userID, err := utils.ValidateToken(tokenString)
	if err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "Authentication error: invalid token", err)
		return
	}

	request, err := utils.DecodeJsonFromRequest[dto.VerifyEmailDTO](w, r.Body)
	if err != nil {
		return
	}

	err = newService.VerifyEmail(userID, request.Email)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to verify email", err)
		return
	}

	response := utils.BuildResponseWithNoBody(w, http.StatusOK, "Verification email sent")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Internal error happened", err)
	}
}

func CheckEmailVerificationCode(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	service := r.Context().Value("service").(interfaces.IService)
	tokenString := r.Header.Get("Authorization")
	tokenString = tokenString[7:] // Remove the "Bearer " prefix
	userID, err := utils.ValidateToken(tokenString)
	if err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "Authentication error: invalid token", err)
		return
	}

	request, err := utils.DecodeJsonFromRequest[dto.CheckEmailVerificationCodeDTO](w, r.Body)
	if err != nil {
		return
	}

	err = service.CheckEmailVerificationCode(userID, request.Code)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to verify code", err)
		return
	}

	user, err := service.FindUser(userID)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "User with provided id does not exist", err)
		return
	}

	zkProof, err := service.GenerateZkProofOfEmailVerification(user, request)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to generate zk proof of email verification", err)
		return
	}

	err = service.SaveProofOfEmailVerification(userID, request.Code, zkProof)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to save proof of the email verification procedure", err)
		return
	}

	response := utils.BuildResponseWithNoBody(w, http.StatusOK, "Your email was successfully verified!")

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Internal error happened", err)
	}
}

func UpdateDisplayNameHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)
	tokenString := r.Header.Get("Authorization")
	tokenString = tokenString[7:] // Remove the "Bearer " prefix
	userID, err := utils.ValidateToken(tokenString)
	if err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "Authentication error: invalid token", err)
		return
	}

	request, err := utils.DecodeJsonFromRequest[dto.UpdateDisplayNameDTO](w, r.Body)
	if err != nil {
		return
	}

	err = newService.UpdateDisplayName(userID, request)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to update display name", err)
		return
	}

	resp := utils.BuildResponseWithNoBody(w, http.StatusOK, "Display name updated successfully")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func GetUsageStats(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodGet) {
		return
	}

	authToken := r.Header.Get("Authorization")
	if authToken == "" {
		utils.HandleError(w, http.StatusUnauthorized, "failed to show client usage statistics", errors.New("missing jwt token"))
		return
	}

	authToken = authToken[7:]
	clientClaims, err := utils.ValidateClientToken(authToken)
	if err != nil {
		utils.HandleError(w, http.StatusUnauthorized, "failed to show client usage statistics", errors.New("jwt token invalid"))
		return
	}

	statRepo := repository.NewStatRepository(db.GetInfluxDBClient())

	now := time.Now()
	firstDayOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	firstDayOfNextMonth := time.Date(firstDayOfMonth.Year(), firstDayOfMonth.Month()+1, 1, 0, 0, 0, 0, time.UTC)
	lastDayOfCurrentMonth := firstDayOfNextMonth.Add(-24 * time.Hour)
	totalDaysInMonth := lastDayOfCurrentMonth.Day()
	totalDaysBeforeNextMonth := totalDaysInMonth - now.Day()

	thirtyDaysStatistic, err := statRepo.GetTotalRequestsInLastXDaysByClient(r.Context(), 30, clientClaims.ClientID)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to get last thrthy days usage statistic", err)
		return
	}

	monthToDateTotal, err := statRepo.GetTotalByDateRangeByClient(r.Context(), firstDayOfMonth, firstDayOfNextMonth, clientClaims.ClientID)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to get month to date usage statistic", err)
		return
	}

	finalResponse := models.UsageStatisticResponse{
		MonthToDate: models.MonthToDateStatistic{
			Month: firstDayOfMonth.Month().String(),
		},
		LastThirtyDaysStatistic: thirtyDaysStatistic,
		MetricType:              "data_transferred",
		UnitOfMeasurement:       "GB",
	}

	if monthToDateTotal > 0 {
		finalResponse.MonthToDate.MonthToDateUsage = monthToDateTotal / 1000000000
		finalResponse.MonthToDate.ForecastedEndOfMonthUsage = (monthToDateTotal / 1000000000) + float64(totalDaysBeforeNextMonth)*thirtyDaysStatistic.Average
	}

	resp := utils.BuildResponse(w, http.StatusOK, "Client usage statistics", finalResponse)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func CheckBackendURI(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	request, err := utils.DecodeJsonFromRequest[dto.CheckBackendURIDTO](w, r.Body)
	if err != nil {
		return
	}

	response, err := newService.CheckBackendURI(request.BackendURI)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to check backend url", err)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: could not encode response into json",
			err,
		)
	}
}

func ResetUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	request, err := utils.DecodeJsonFromRequest[dto.PasswordResetDTO](w, r.Body)
	if err != nil {
		return
	}

	user, err := newService.FindUserForUsername(request.Username)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "User not found for username", err)
		return
	}

	err = newService.VerifyUserEmailProof(&user, request.EmailVerificationCode)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to check if the provided email was verified", err)
		return
	}

	token, err := newService.GeneratePasswordResetToken()
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to generate the token", err)
		return
	}

	err = newService.SendPasswordResetToken(token, &user, request.Email)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to send the reset link", err)
		return
	}

	err = newService.SavePasswordResetToken(token, &user)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to save the password reset token", err)
		return
	}

	resp := utils.BuildResponseWithNoBody(
		w,
		http.StatusOK,
		"We have sent email with a reset password link to your email address",
	)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		utils.HandleError(
			w,
			http.StatusInternalServerError,
			"Internal error: failed to encode response into json",
			err,
		)
	}
}

func VerifyPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodGet) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	err := r.ParseForm()
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Error while parsing the query parameters", err)
		return
	}

	token := r.Form.Get("token")

	tokenData, err := newService.GetPasswordResetTokenData(token)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Token does not exist", err)
		return
	}

	err = newService.ValidatePasswordResetTokenData(tokenData)
	if err != nil {
		utils.HandleError(w, http.StatusBadRequest, "Failed to validate password reset token", err)
		return
	}

	ServeFileHandlerWithArgs(
		w,
		r,
		"assets-v1/templates/src/pages/user_portal/password_reset/input-new-password.html",
		map[string]interface{}{
			"USERNAME": tokenData.Username,
		},
	)
}

func UpdateUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	if !validateHttpMethod(w, r.Method, http.MethodPost) {
		return
	}

	newService := r.Context().Value("service").(interfaces.IService)

	request, err := utils.DecodeJsonFromRequest[dto.UpdatePasswordDTO](w, r.Body)
	if err != nil {
		return
	}

	err = newService.UpdateUserPassword(request)
	if err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to update user's password", err)
		return
	}

	response := utils.BuildResponseWithNoBody(
		w, http.StatusCreated, "Your password was updated successfully!",
	)
	if err := json.NewEncoder(w).Encode(&response); err != nil {
		utils.HandleError(w, http.StatusInternalServerError, "Failed to encode the response", err)
	}
}

func validateHttpMethod(w http.ResponseWriter, actualMethod string, expectedMethod string) bool {
	if actualMethod != expectedMethod {
		errorMessage := fmt.Sprintf("Invalid http method. Expected %s", expectedMethod)
		utils.HandleError(
			w,
			http.StatusMethodNotAllowed,
			errorMessage,
			fmt.Errorf(errorMessage),
		)
		return false
	}

	return true
}
