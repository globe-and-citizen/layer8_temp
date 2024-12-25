package repository

import (
	"fmt"
	serverModels "globe-and-citizen/layer8/server/models"
	"globe-and-citizen/layer8/server/resource_server/dto"
	interfaces "globe-and-citizen/layer8/server/resource_server/interfaces"
	"globe-and-citizen/layer8/server/resource_server/models"
	"globe-and-citizen/layer8/server/resource_server/utils"
	"strconv"
	"strings"
	"time"
)

type MemoryRepository struct {
	storage                 map[string]map[string]string
	byteStorage             map[string][]byte
	verificationDataStorage map[string]models.EmailVerificationData
}

func NewMemoryRepository() interfaces.IRepository {
	return &MemoryRepository{
		storage:                 make(map[string]map[string]string),
		byteStorage:             make(map[string][]byte),
		verificationDataStorage: make(map[string]models.EmailVerificationData),
	}
}

func (r *MemoryRepository) RegisterUser(req dto.RegisterUserDTO, hashedPassword string, salt string) error {
	rmSalt := utils.GenerateRandomSalt(utils.SaltSize)
	HashedAndSaltedPass := utils.SaltAndHashPassword(req.Password, rmSalt) // what if two user's use the same password?
	userID := fmt.Sprintf("%d", len(r.storage))
	key := fmt.Sprintf("%s:%s", req.Username, req.Password)
	r.storage[key] = map[string]string{
		"user_id":           userID,
		"username":          req.Username,
		"password":          HashedAndSaltedPass,
		"first_name":        req.FirstName,
		"last_name":         req.LastName,
		"country":           req.Country,
		"display_name":      req.DisplayName,
		"salt":              rmSalt,
		"email_verified":    "false",
		"email_proof":       "",
		"verification_code": "",
	}
	r.storage[req.Username] = map[string]string{
		"salt":     rmSalt,
		"password": key,
	}
	r.storage[userID] = map[string]string{
		"password": key,
	}
	return nil
}

func (r *MemoryRepository) RegisterClient(client models.Client) error {
	if _, ok := r.storage[client.Username]; ok {
		return fmt.Errorf("Client username already registered.")
	}
	clientUUID := utils.GenerateUUID()
	clientSecret := utils.GenerateSecret(utils.SecretSize)
	r.storage[client.Username] = map[string]string{
		"id":           clientUUID,
		"secret":       clientSecret,
		"redirect_uri": client.RedirectURI,
		"backend_uri":  client.BackendURI,
		"username":     client.Username,
		"password":     client.Password,
	}
	return nil
}

func (r *MemoryRepository) getUserData(userId uint) (map[string]string, error) {
	userKeyData, ok := r.storage[strconv.Itoa(int(userId))]
	if !ok {
		return nil, fmt.Errorf("user not found for id %d", userId)
	}
	key := userKeyData["password"]

	userData, ok := r.storage[key]
	if !ok {
		return nil, fmt.Errorf("user not found for key %s", key)
	}

	return userData, nil
}

func (r *MemoryRepository) FindUser(userId uint) (models.User, error) {
	userData, e := r.getUserData(userId)
	if e != nil {
		return models.User{}, e
	}

	return models.User{
		ID:               userId,
		Username:         userData["username"],
		Password:         userData["password"],
		FirstName:        userData["first_name"],
		LastName:         userData["last_name"],
		Salt:             userData["salt"],
		EmailProof:       []byte(userData["email_proof"]),
		VerificationCode: userData["verification_code"],
	}, nil
}

func (r *MemoryRepository) GetClientData(clientName string) (models.Client, error) {
	if _, ok := r.storage[clientName]; !ok {
		return models.Client{}, fmt.Errorf("client not found")
	}
	client := models.Client{
		ID:          r.storage[clientName]["id"],
		Secret:      r.storage[clientName]["secret"],
		Name:        clientName,
		RedirectURI: r.storage[clientName]["redirect_uri"],
	}
	return client, nil
}

func (r *MemoryRepository) LoginPreCheckUser(req dto.LoginPrecheckDTO) (string, string, error) {
	if _, ok := r.storage[req.Username]["salt"]; !ok {
		return "", "", fmt.Errorf("salt not found for specified user")
	}
	return req.Username, r.storage[req.Username]["salt"], nil
}

func (r *MemoryRepository) LoginUser(req dto.LoginUserDTO) (models.User, error) {
	key := fmt.Sprintf("%s:%s", req.Username, req.Password)
	if _, ok := r.storage[key]; !ok {
		return models.User{}, fmt.Errorf("user not found")
	}
	if r.storage[key]["username"] != req.Username {
		return models.User{}, fmt.Errorf("invalid username")
	}
	UserID := r.storage[key]["user_id"]
	userIdUint, err := strconv.ParseUint(UserID, 10, 32)
	if err != nil {
		return models.User{}, err
	}
	user := models.User{
		ID:        uint(userIdUint),
		Username:  r.storage[key]["username"],
		Password:  r.storage[key]["password"],
		FirstName: r.storage[key]["first_name"],
		LastName:  r.storage[key]["last_name"],
		Salt:      r.storage[key]["salt"],
	}
	return user, nil
}

func (r *MemoryRepository) LoginClient(req dto.LoginClientDTO) (models.Client, error) {
	if _, ok := r.storage[req.Username]; !ok {
		return models.Client{}, fmt.Errorf("user not found")
	}

	if r.storage[req.Username]["password"] != req.Password {
		return models.Client{}, fmt.Errorf("invalid password")
	}

	client := models.Client{
		ID:          r.storage[req.Username]["id"],
		Secret:      r.storage[req.Username]["secret"],
		RedirectURI: r.storage[req.Username]["redirect_uri"],
		Username:    r.storage[req.Username]["username"],
	}

	fmt.Println("client: ", client)
	return client, nil
}

func (r *MemoryRepository) ProfileUser(userID uint) (models.User, []models.UserMetadata, error) {
	if _, ok := r.storage[fmt.Sprintf("%d", userID)]; !ok {
		return models.User{}, []models.UserMetadata{}, fmt.Errorf("user not found")
	}
	password := r.storage[fmt.Sprintf("%d", userID)]["password"]
	user := models.User{
		ID:        userID,
		Username:  r.storage[password]["username"],
		Password:  r.storage[password]["password"],
		FirstName: r.storage[password]["first_name"],
		LastName:  r.storage[password]["last_name"],
		Salt:      r.storage[password]["salt"],
	}
	userMetadata := []models.UserMetadata{
		{
			Key:   "display_name",
			Value: r.storage[password]["display_name"],
		},
		{
			Key:   "country",
			Value: r.storage[password]["country"],
		},
		{
			Key:   "email_verified",
			Value: r.storage[password]["email_verified"],
		},
	}
	return user, userMetadata, nil
}

func (r *MemoryRepository) ProfileClient(username string) (models.Client, error) {
	if _, ok := r.storage[username]; !ok {
		return models.Client{}, fmt.Errorf("user not found")
	}

	client := models.Client{
		ID:          r.storage[username]["id"],
		Secret:      r.storage[username]["secret"],
		RedirectURI: r.storage[username]["redirect_uri"],
		Username:    username,
	}
	return client, nil
}

func (r *MemoryRepository) SaveProofOfEmailVerification(
	userId uint, verificationCode string, emailProof []byte, zkKeyPairId uint,
) error {
	userData, e := r.getUserData(userId)
	if e != nil {
		return e
	}

	userData["verification_code"] = verificationCode
	userData["email_proof"] = string(emailProof)
	userData["email_verified"] = "true"

	delete(r.verificationDataStorage, strconv.Itoa(int(userId)))

	return nil
}

func (r *MemoryRepository) SaveEmailVerificationData(data models.EmailVerificationData) error {
	r.verificationDataStorage[strconv.Itoa(int(data.UserId))] = data
	return nil
}

func (r *MemoryRepository) GetEmailVerificationData(userId uint) (models.EmailVerificationData, error) {
	data, ok := r.verificationDataStorage[strconv.Itoa(int(userId))]
	if !ok {
		return models.EmailVerificationData{},
			fmt.Errorf("verification data not found for user with id %d", userId)
	}

	return data, nil
}

func (r *MemoryRepository) UpdateDisplayName(userID uint, req dto.UpdateDisplayNameDTO) error {
	if _, ok := r.storage[fmt.Sprintf("%d", userID)]; !ok {
		return fmt.Errorf("user not found")
	}
	password := r.storage[fmt.Sprintf("%d", userID)]["password"]
	r.storage[password]["display_name"] = req.DisplayName
	return nil
}

// Oauth methods
func (r *MemoryRepository) LoginUserPrecheck(username string) (string, error) {
	if _, ok := r.storage[username]; !ok {
		fmt.Println("user not found while using LoginUserPrecheck")
		return "", fmt.Errorf("user not found")
	}
	return r.storage[username]["salt"], nil
}

func (r *MemoryRepository) LoginPreCheckClient(req dto.LoginPrecheckDTO) (string, string, error) {

	return "r.storage[username]['salt']", "second string...", nil
}

func (r *MemoryRepository) GetUser(username string) (*serverModels.User, error) {
	if _, ok := r.storage[username]; !ok {
		fmt.Println("user not found while using GetUser")
		return &serverModels.User{}, fmt.Errorf("user not found")
	}
	password := r.storage[username]["password"]
	userID := r.storage[password]["user_id"]
	userIdInt, err := strconv.Atoi(userID)
	if err != nil {
		return &serverModels.User{}, err
	}
	user := serverModels.User{
		ID:        uint(userIdInt),
		Email:     r.storage[password]["email"],
		Username:  r.storage[password]["username"],
		FirstName: r.storage[password]["first_name"],
		LastName:  r.storage[password]["last_name"],
		Password:  r.storage[password]["password"],
		Salt:      r.storage[username]["salt"],
	}
	return &user, nil
}

func (r *MemoryRepository) GetUserByID(id int64) (*serverModels.User, error) {
	if _, ok := r.storage[fmt.Sprintf("%d", id)]; !ok {
		fmt.Println("user not found while using GetUserByID")
		return &serverModels.User{}, fmt.Errorf("user not found")
	}
	password := r.storage[fmt.Sprintf("%d", id)]["password"]
	user := serverModels.User{
		ID:        uint(id),
		Email:     r.storage[password]["email"],
		Username:  r.storage[password]["username"],
		FirstName: r.storage[password]["first_name"],
		LastName:  r.storage[password]["last_name"],
		Password:  r.storage[password]["password"],
		Salt:      r.storage[password]["salt"],
	}
	return &user, nil
}

func (r *MemoryRepository) GetUserMetadata(userID int64, key string) (*serverModels.UserMetadata, error) {
	if _, ok := r.storage[fmt.Sprintf("%d", userID)]; !ok {
		fmt.Println("user not found while using GetUserMetadata")
		return &serverModels.UserMetadata{}, fmt.Errorf("user not found")
	}
	password := r.storage[fmt.Sprintf("%d", userID)]["password"]
	userMetadata := serverModels.UserMetadata{
		Key:   key,
		Value: r.storage[password][key],
	}
	return &userMetadata, nil
}

func (r *MemoryRepository) SetClient(client *serverModels.Client) error {
	r.storage[client.ID] = map[string]string{
		"id":           client.ID,
		"secret":       client.Secret,
		"name":         client.Name,
		"redirect_uri": client.RedirectURI,
		"backend_uri":  client.BackendURI,
		"username":     client.Username,
		"password":     client.Password,
		"salt":         client.Salt,
	}
	return nil
}

func (r *MemoryRepository) GetClient(id string) (*serverModels.Client, error) {
	if strings.Contains(id, ":") {
		id = id[strings.LastIndex(id, ":")+1:]
		// fmt.Println("ID check:", id)
	}
	if _, ok := r.storage[id]; !ok {
		fmt.Println("client not found while using GetClient")
		return &serverModels.Client{}, fmt.Errorf("client not found")
	}
	client := serverModels.Client{
		ID:          r.storage[id]["id"],
		Secret:      r.storage[id]["secret"],
		Name:        r.storage[id]["name"],
		RedirectURI: r.storage[id]["redirect_uri"],
	}
	return &client, nil
}

func (r *MemoryRepository) GetClientDataByBackendURL(backendUrl string) (models.Client, error) {
	for _, data := range r.storage {
		backend, ok := data["backend_uri"]
		if ok && backend == backendUrl {
			return models.Client{
				ID:          data["id"],
				Secret:      data["secret"],
				Name:        data["name"],
				RedirectURI: data["redirect_uri"],
				BackendURI:  backend,
				Username:    data["username"],
				Password:    data["password"],
				Salt:        data["salt"],
			}, nil
		}
	}

	fmt.Printf("client not found for backend url %s\n", backendUrl)
	return models.Client{}, fmt.Errorf("client not found")
}

func (r *MemoryRepository) SetTTL(key string, value []byte, ttl time.Duration) error {
	r.byteStorage[key] = value
	go func() {
		time.Sleep(ttl)
		delete(r.storage, key)
	}()
	return nil
}

func (r *MemoryRepository) GetTTL(key string) ([]byte, error) {
	if _, ok := r.byteStorage[key]; !ok {
		fmt.Println("key not found while using GetTTL")
		return nil, fmt.Errorf("key not found")
	}
	return r.byteStorage[key], nil
}

func (r *MemoryRepository) IsBackendURIExists(backendURL string) (bool, error) {
	for _, data := range r.storage {
		backend, ok := data["backend_uri"]
		if ok && backend == backendURL {
			return true, nil
		}
	}
	return false, nil
}

func (r *MemoryRepository) SaveZkSnarksKeyPair(keyPair models.ZkSnarksKeyPair) (uint, error) {
	return 0, nil
}

func (r *MemoryRepository) GetLatestZkSnarksKeys() (models.ZkSnarksKeyPair, error) {
	return models.ZkSnarksKeyPair{}, nil
}

func (r *MemoryRepository) GetUserForUsername(username string) (models.User, error) {
	return models.User{}, nil
}

func (r *MemoryRepository) UpdateUserPassword(username string, password string) error {
	return nil
}

func (r *MemoryRepository) RegisterPrecheckUser(req dto.RegisterUserPrecheckDTO, salt string, iterCount int) (string, int, error) {
	if _, exists := r.storage[req.Username]; exists {
		return "", 0, fmt.Errorf("user already exists: %s", req.Username)
	}

	userID := fmt.Sprintf("%d", len(r.storage))

	key := fmt.Sprintf("%s:%d", req.Username, iterCount)

	r.storage[key] = map[string]string{
		"user_id":        userID,
		"username":       req.Username,
		"salt":           salt,
		"iterationCount": "",
	}

	r.storage[req.Username] = map[string]string{
		"salt":     salt,
		"password": key,
	}
	r.storage[userID] = map[string]string{
		"password": key,
	}

	return salt, iterCount, nil
}