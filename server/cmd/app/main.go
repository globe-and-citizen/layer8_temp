package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint"
	"globe-and-citizen/layer8/server/config"
	"globe-and-citizen/layer8/server/handlers"
	"globe-and-citizen/layer8/server/opentelemetry"
	"globe-and-citizen/layer8/server/resource_server/db"
	"globe-and-citizen/layer8/server/resource_server/emails/sender"
	"globe-and-citizen/layer8/server/resource_server/emails/verification"
	"globe-and-citizen/layer8/server/resource_server/emails/verification/code"
	"globe-and-citizen/layer8/server/resource_server/emails/verification/zk"
	"globe-and-citizen/layer8/server/resource_server/models"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	Ctl "globe-and-citizen/layer8/server/resource_server/controller"
	"globe-and-citizen/layer8/server/resource_server/dto"
	"globe-and-citizen/layer8/server/resource_server/interfaces"
	"globe-and-citizen/layer8/server/resource_server/utils"

	oauthRepo "globe-and-citizen/layer8/server/internals/repository"

	rsRepo "globe-and-citizen/layer8/server/resource_server/repository"

	svc "globe-and-citizen/layer8/server/resource_server/service" // there are two services

	oauthSvc "globe-and-citizen/layer8/server/internals/service" // there are two services

	"github.com/joho/godotenv"
)

// go:embed dist
var StaticFiles embed.FS

var workingDirectory string

func getPwd() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	workingDirectory = dir
}

func main() {
	// Use flags to set the port
	port := flag.Int("port", 8080, "Port to run the server on")
	jwtKey := flag.String("jwtKey", "secret", "Key to sign JWT tokens")
	MpKey := flag.String("MpKey", "secret", "Key to sign mpJWT tokens")
	UpKey := flag.String("UpKey", "secret", "Key to sign upJWT tokens")
	ProxyURL := flag.String("ProxyURL", "http://localhost:5001", "URL to populate go HTML templates")
	InMemoryDb := flag.Bool(
		"InMemoryDb",
		false,
		"Defines whether or not to use the in-memory database implementation")

	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	if err := opentelemetry.NewMeter(context.Background()); err != nil {
		log.Fatalf("Failed to create meter: %v", err)
	}

	db.InitInfluxDBClient()

	var resourceRepository interfaces.IRepository
	var oauthService *oauthSvc.Service

	if *InMemoryDb {
		os.Setenv("SERVER_PORT", strconv.Itoa(*port))
		os.Setenv("JWT_SECRET_KEY", *jwtKey)
		os.Setenv("MP_123_SECRET_KEY", *MpKey)
		os.Setenv("UP_999_SECRET_KEY", *UpKey)
		os.Setenv("PROXY_URL", *ProxyURL)

		resourceRepository = rsRepo.NewMemoryRepository()
		service := svc.NewService(resourceRepository, &verification.EmailVerifier{}, &zk.ProofProcessor{})
		service.RegisterUser(dto.RegisterUserDTO{
			Username:    "tester",
			FirstName:   "Test",
			LastName:    "User",
			Password:    "12341234",
			Country:     "Antarctica",
			DisplayName: "test_user_mem",
		})

		oauthService = &oauthSvc.Service{Repo: resourceRepository}

		fmt.Println("Running app with in-memory repository")
	} else {
		// If the user has set a database user or password, init the database
		if os.Getenv("DB_USER") != "" || os.Getenv("DB_PASSWORD") != "" {
			config.InitDB()
		}

		resourceRepository = rsRepo.NewRepository(config.DB)
		oauthService = &oauthSvc.Service{Repo: oauthRepo.NewOauthRepository(config.DB)}

		fmt.Println("Running the app with postgres repository")
	}

	adminEmailAddress := fmt.Sprintf(
		"%s@%s",
		os.Getenv("LAYER8_EMAIL_USERNAME"),
		os.Getenv("LAYER8_EMAIL_DOMAIN"),
	)

	verificationCodeValidityDuration, e :=
		time.ParseDuration(os.Getenv("VERIFICATION_CODE_VALIDITY_DURATION"))
	if e != nil {
		log.Fatalf("error parsing verification code validity duration: %e", e)
	}

	emailVerifier := verification.NewEmailVerifier(
		adminEmailAddress,
		sender.NewMailerSendService(
			os.Getenv("MAILER_SEND_API_KEY"),
			os.Getenv("MAILER_SEND_TEMPLATE_ID"),
		),
		code.NewMIMCCodeGenerator(),
		verificationCodeValidityDuration,
		time.Now,
	)

	generateNewKeys, err := strconv.ParseBool(os.Getenv("GENERATE_NEW_ZK_SNARKS_KEYS"))
	if err != nil {
		log.Fatalf("Error while parsing GENERATE_NEW_ZK_SNARKS_KEYS flag: %e", err)
	}

	var cs constraint.ConstraintSystem
	var zkKeyPairId uint
	var provingKey groth16.ProvingKey
	var verifyingKey groth16.VerifyingKey

	if generateNewKeys {
		cs, provingKey, verifyingKey = zk.RunZkSnarksSetup()

		zkKeyPairId, err = resourceRepository.SaveZkSnarksKeyPair(
			models.ZkSnarksKeyPair{
				ProvingKey:   utils.WriteBytes(provingKey),
				VerifyingKey: utils.WriteBytes(verifyingKey),
			},
		)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		zkSnarksKeyPair, err := resourceRepository.GetLatestZkSnarksKeys()
		if err != nil {
			log.Fatalf("Error while reading zk-snarks keys from the database: %e", err)
		}

		cs = zk.GenerateConstraintSystem()
		zkKeyPairId = zkSnarksKeyPair.ID

		// Empty proving key initialised with elliptic curve id
		provingKey = groth16.NewProvingKey(ecc.BN254)
		// Deserialize proving key representation bytes from db into the provingKey object
		utils.ReadBytes[groth16.ProvingKey](provingKey, zkSnarksKeyPair.ProvingKey)

		// Empty verifying key initialised with elliptic curve id
		verifyingKey = groth16.NewVerifyingKey(ecc.BN254)
		// Deserialize verifying key representation bytes from db into the verifyingKey object
		utils.ReadBytes[groth16.VerifyingKey](verifyingKey, zkSnarksKeyPair.VerifyingKey)
	}

	proofProcessor := zk.NewProofProcessor(cs, zkKeyPairId, provingKey, verifyingKey)

	// Run server (which never returns)
	Server(
		svc.NewService(resourceRepository, emailVerifier, proofProcessor),
		oauthService,
	)
}

func Server(resourceService interfaces.IService, oauthService *oauthSvc.Service) {
	port := os.Getenv("SERVER_PORT")

	getPwd()

	authenticationHandler := handlers.NewAuthenticationHandler(
		oauthService,
		utils.ParseHTML,
	)

	authorizationHandler := handlers.NewAuthorizationHandler(
		oauthService,
		utils.ParseHTML,
	)

	server := http.Server{
		Addr: fmt.Sprintf(":%s", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "*")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			r = r.WithContext(context.WithValue(r.Context(), "Oauthservice", oauthService))
			r = r.WithContext(context.WithValue(r.Context(), "service", resourceService))

			staticFS, _ := fs.Sub(StaticFiles, "dist")
			httpFS := http.FileServer(http.FS(staticFS))

			if r.Header.Get("up-JWT") != "" {
				handlers.Tunnel(w, r)
				return
			}

			switch path := r.URL.Path; {

			// Authorization Server endpoints
			case path == "/login":
				authenticationHandler.Login(w, r)
			case path == "/authorize":
				authorizationHandler.Authorize(w, r)
			case path == "/error":
				authorizationHandler.Error(w, r)
			case path == "/api/oauth":
				authorizationHandler.OAuthToken(w, r)
			case path == "/api/user":
				handlers.UserInfo(w, r)
			case strings.HasPrefix(path, "/assets-v1"):
				http.StripPrefix("/assets-v1", http.FileServer(http.Dir("./assets-v1"))).ServeHTTP(w, r)

			// Resource Server endpoints
			case path == "/":
				Ctl.IndexHandler(w, r)
			case path == "/user":
				Ctl.UserHandler(w, r)
			case path == "/user-login-page":
				Ctl.LoginUserPage(w, r)
			case path == "/user-register-page":
				Ctl.RegisterUserPage(w, r)
			case path == "/input-your-email-page":
				Ctl.InputYourEmailPage(w, r)
			case path == "/input-verification-code-page":
				Ctl.InputVerificationCodePage(w, r)
			case path == "/client-register-page":
				Ctl.ClientHandler(w, r)
			case path == "/client-login-page":
				Ctl.LoginClientPage(w, r)
			case path == "/client-profile":
				Ctl.ClientProfilePage(w, r)
			case path == "/reset-password-page":
				Ctl.ResetPasswordPage(w, r)
			case path == "/api/v1/register-user":
				Ctl.RegisterUserHandler(w, r)
			case path == "/api/v1/register-user-precheck":
				Ctl.RegisterUserPrecheck(w, r)
			case path == "/api/v1/register-client":
				Ctl.RegisterClientHandler(w, r)
			case path == "/api/v1/getClient":
				Ctl.GetClientData(w, r)
			case path == "/api/v1/login-precheck":
				Ctl.LoginPrecheckHandler(w, r)
			case path == "/api/v1/login-user":
				Ctl.LoginUserHandler(w, r)
			case path == "/api/v1/login-client":
				Ctl.LoginClientHandler(w, r) // Login Client
			case path == "/api/v1/profile":
				Ctl.ProfileHandler(w, r)
			case path == "/api/v1/client-profile":
				Ctl.ClientProfileHandler(w, r)
			case path == "/api/v1/verify-email":
				Ctl.VerifyEmailHandler(w, r)
			case path == "/api/v1/check-email-verification-code":
				Ctl.CheckEmailVerificationCode(w, r)
			case path == "/api/v1/change-display-name":
				Ctl.UpdateDisplayNameHandler(w, r)
			case path == "/api/v1/usage-stats":
				Ctl.GetUsageStats(w, r)
			case path == "/api/v1/check-backend-uri":
				Ctl.CheckBackendURI(w, r)
			case path == "/api/v1/reset-password":
				Ctl.ResetPasswordHandler(w, r)
			case path == "/favicon.ico":
				faviconPath := workingDirectory + "/dist/favicon.ico"
				http.ServeFile(w, r, faviconPath)
			case strings.HasPrefix(path, "/assets/"):
				httpFS.ServeHTTP(w, r)

			// Proxy Server endpoints
			case path == "/init-tunnel":
				handlers.InitTunnel(w, r)
			case path == "/error":
				handlers.TestError(w, r)
				// TODO: For later, to be discussed more
				// case path == "/tunnel":
				// 	handlers.Tunnel(w, r)
			}
		}),
	}
	log.Printf("Starting server on port %s...", port)
	log.Fatal(server.ListenAndServe())
}
