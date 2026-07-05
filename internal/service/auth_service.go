package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"task-system/internal/models"
	"task-system/internal/repository"
	"task-system/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo        *repository.UserRepository
	sessionRepo *repository.SessionRepository
}

func NewAuthService(repo *repository.UserRepository, sessionRepo *repository.SessionRepository) *AuthService {
	return &AuthService{
		repo:        repo,
		sessionRepo: sessionRepo,
	}
}

// Register contains signup business logic: normalizing email, hashing password, and saving to database
func (s *AuthService) Register(name, email, password string) (*models.User, error) {
	cleanedEmail := strings.ToLower(strings.TrimSpace(email))

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:     strings.TrimSpace(name),
		Email:    cleanedEmail,
		Password: string(passwordHash),
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

// Signin authenticates user email and password
func (s *AuthService) Signin(email, password string) (*models.User, error) {
	cleanedEmail := strings.ToLower(strings.TrimSpace(email))

	user, err := s.repo.GetByEmail(cleanedEmail)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, err
	}

	return user, nil
}

// CreateSession generates a new Access/Refresh token pair, invalidates old sessions, and persists state to PostgreSQL
func (s *AuthService) CreateSession(user *models.User) (string, string, error) {
	// 1. Enforce single-login constraint: clear old database sessions
	err := s.clearUserSession(user.ID)
	if err != nil {
		return "", "", err
	}

	// 2. Generate a secure random string for the raw Refresh Token secret
	rawSecret, err := generateRandomToken(32)
	if err != nil {
		return "", "", err
	}

	// 3. Encrypt the Refresh Token using symmetric AES-256 GCM for secure database storage
	encryptedToken, err := utils.Encrypt(rawSecret)
	if err != nil {
		return "", "", err
	}

	// 4. Save new session to PostgreSQL (30 days TTL)
	sessionExpiry := time.Now().Add(30 * 24 * time.Hour).Unix()
	session := &models.Session{
		UserID:    user.ID,
		Token:     encryptedToken,
		ExpiresAt: sessionExpiry,
	}
	err = s.sessionRepo.Create(session)
	if err != nil {
		return "", "", err
	}

	// 5. Generate stateless JWT Access Token containing user_id and session_id
	accessToken, err := utils.GenerateAccessToken(user.ID, session.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	// Return compound Refresh Token formatted as "userID.rawSecret"
	compoundRefreshToken := user.ID + "." + rawSecret
	return accessToken, compoundRefreshToken, nil
}

// RefreshSession validates an encrypted compound refresh token and issues rotated tokens
func (s *AuthService) RefreshSession(compoundRefreshToken string) (string, string, error) {
	// Split compound token into userID and raw secret
	parts := strings.Split(compoundRefreshToken, ".")
	if len(parts) != 2 {
		return "", "", errors.New("invalid refresh token format")
	}

	userID := parts[0]
	rawSecret := parts[1]

	// Fetch active session by userID
	session, err := s.sessionRepo.GetByUserID(userID)
	if err != nil {
		return "", "", err
	}

	// Decrypt the stored token
	decryptedSecret, err := utils.Decrypt(session.Token)
	if err != nil || decryptedSecret != rawSecret {
		return "", "", errors.New("invalid refresh token")
	}

	// Verify expiration
	if session.ExpiresAt < time.Now().Unix() {
		_ = s.sessionRepo.DeleteByUserID(userID)
		return "", "", errors.New("session expired")
	}

	// Fetch user details
	user, err := s.repo.GetByID(session.UserID)
	if err != nil {
		return "", "", err
	}

	// Token Rotation: Generate a new raw secret
	newRawSecret, err := generateRandomToken(32)
	if err != nil {
		return "", "", err
	}

	// Encrypt the new secret
	newEncryptedToken, err := utils.Encrypt(newRawSecret)
	if err != nil {
		return "", "", err
	}

	// Extend session expiration date to 30 days from now (Sliding Expiration)
	newExpiresAt := time.Now().Add(30 * 24 * time.Hour).Unix()

	// Update PostgreSQL with the new encrypted token and new expiration date
	err = s.sessionRepo.UpdateSession(session.ID, newEncryptedToken, newExpiresAt)
	if err != nil {
		return "", "", err
	}

	// Generate a new JWT access token using the same active session.ID
	newAccessToken, err := utils.GenerateAccessToken(user.ID, session.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	// Return the new rotated Access and Refresh tokens
	newCompoundRefreshToken := user.ID + "." + newRawSecret
	return newAccessToken, newCompoundRefreshToken, nil
}

// InvalidateSession clears the session from PostgreSQL (Logout)
func (s *AuthService) InvalidateSession(compoundRefreshToken string, userID string) error {
	return s.clearUserSession(userID)
}

// clearUserSession is a common helper that deletes all sessions of a user in Postgres
func (s *AuthService) clearUserSession(userID string) error {
	// Delete PostgreSQL session records
	return s.sessionRepo.DeleteByUserID(userID)
}

// Helper: Generates a cryptographically secure random hex string
func generateRandomToken(bytesLen int) (string, error) {
	bytes := make([]byte, bytesLen)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}