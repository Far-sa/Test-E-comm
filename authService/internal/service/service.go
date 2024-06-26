package service

import (
	"auth-svc/internal/param"
	"auth-svc/internal/ports"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
	//"github.com/dgrijalva/jwt-go"
)

type Config struct {
	JwtSignKey                     string
	AccessTokenSubject             string
	RefreshTokenSubject            string
	AccessTokenExpirationDuration  time.Duration
	RefreshTokenExpirationDuration time.Duration
}

// TODO: add to config
const (
	JwtSignKey                     = "jwt-secret"
	AccessTokenSubject             = "at"
	RefreshTokenSubject            = "rt"
	AccessTokenExpirationDuration  = time.Hour * 24
	RefreshTokenExpirationDuration = time.Hour * 24 * 7
)

type authService struct {
	config        ports.Config
	authRepo      ports.AuthRepository
	messageBroker ports.EventPublisher
}

// NewTokenHandler creates a new TokenHandler with the given authService
func NewAuthService(config ports.Config, authRepo ports.AuthRepository, event ports.EventPublisher) authService {
	return authService{config: config, authRepo: authRepo, messageBroker: event}
}

func (s authService) Login(ctx context.Context, req param.LoginRequest) (param.LoginResponse, error) {

	//! Increase the timeout to 60 seconds for testing purposes
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second) // Adjust the timeout as needed
	defer cancel()

	err := s.publishLoginRequest(ctx, req)
	if err != nil {
		log.Printf("Failed to publish login request: %v", err)
		return param.LoginResponse{}, fmt.Errorf("failed to publish login request: %w", err)
	}

	log.Println("Waiting for user service response")

	response, err := s.waitForUserServiceResponse(ctx, req.Email)
	if err != nil {
		log.Printf("Failed to get user service response: %v", err)
		return param.LoginResponse{}, fmt.Errorf("failed to get user service response: %w", err)
	}

	log.Printf("Received response from user service: %+v", response)

	accessToken, err := s.createAccessToken(response)
	if err != nil {
		return param.LoginResponse{}, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken, err := s.refreshAccessToken(response)
	if err != nil {
		return param.LoginResponse{}, fmt.Errorf("failed to create refresh token: %w", err)
	}

	if err := s.authRepo.StoreToken(response.User.ID, accessToken, time.Now().Add(72*time.Hour)); err != nil {
		//! TODO: log error
		fmt.Println("Error storing token:", err)
	}

	return param.LoginResponse{
		User:   param.UserInfo{ID: response.User.ID, Email: response.User.Email},
		Tokens: param.Tokens{AccessToken: accessToken, RefreshToken: refreshToken},
	}, nil

}

func (s authService) publishLoginRequest(ctx context.Context, req param.LoginRequest) error {

	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal login request: %w", err)
	}

	if err := s.messageBroker.Publish(ctx, "auth_exchange", "login", amqp.Publishing{
		ContentType:   "application/json",
		DeliveryMode:  amqp.Persistent,
		Body:          data,
		CorrelationId: uuid.NewString(),
	}); err != nil {
		return fmt.Errorf("failed to publish user to auth-service: %w", err)
	}

	log.Println("User event published successfully")
	return nil
}

// !!
func (s *authService) waitForUserServiceResponse(ctx context.Context, email string) (param.LoginResponse, error) {

	// if err := s.messageBroker.DeclareExchange("auth_exchange", "direct"); err != nil {
	// 	return param.LoginResponse{}, fmt.Errorf("failed to declare exchange: %w", err)
	// }

	// queue, err := s.messageBroker.CreateQueue("", true, true) //! use a unique, non-durable queue
	// if err != nil {
	// 	return param.LoginResponse{}, fmt.Errorf("failed to create queue: %w", err)
	// }

	queueName := "user_service_responses_" + uuid.NewString()
	queue, err := s.messageBroker.CreateQueue(queueName, false, false)
	if err != nil {
		return param.LoginResponse{}, fmt.Errorf("failed to create queue: %w", err) // Propagate error
	}
	defer s.messageBroker.DeleteQueue(queueName)

	if err := s.messageBroker.CreateBinding(queue.Name, "user_response", "auth_exchange"); err != nil {
		return param.LoginResponse{}, fmt.Errorf("failed to bind queue: %w", err) // Propagate error

	}

	responseChan, errorChan := s.consumeMessages(ctx, queue.Name, email)

	select {
	case <-ctx.Done():
		log.Println("Context canceled while waiting for user service response")
		return param.LoginResponse{}, ctx.Err()
	case response := <-responseChan:
		log.Println("Received user service response")
		return response, nil
	case err := <-errorChan:
		log.Printf("Error occurred while waiting for user service response: %v", err)
		return param.LoginResponse{}, err
	}
}

func (s *authService) consumeMessages(ctx context.Context, queueName, email string) (<-chan param.LoginResponse, <-chan error) {
	responseChan := make(chan param.LoginResponse)
	errorChan := make(chan error)

	consumerTag := "auth_service_" + uuid.NewString()

	go func() {
		defer close(responseChan)
		defer close(errorChan)

		msgs, err := s.messageBroker.Consume(queueName, consumerTag, false)
		if err != nil {
			errorChan <- fmt.Errorf("failed to consume messages: %w", err)
			return
		}

		for {
			select {
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			case d, ok := <-msgs:
				if !ok {
					errorChan <- fmt.Errorf("message channel closed")
					return
				}

				var response param.LoginResponse
				if err := json.Unmarshal(d.Body, &response); err != nil {
					errorChan <- fmt.Errorf("failed to unmarshal response: %w", err)
					return
				}

				// TODO messages are acknowledged only if the email matches. This could potentially lead to unacknowledged messages
				//! Potential Dead Letter Queue
				if response.User.Email == email {
					d.Ack(false)
					responseChan <- response
					return
				} else {
					d.Nack(false, true) // Requeue the message
				}
			}
		}
	}()
	return responseChan, errorChan
}

//!!@@

func UnmarshalData(data []byte) (param.UserResponse, error) {
	var user param.UserResponse
	err := json.Unmarshal(data, &user)
	return user, err
}

func ComparePassword(hashedPassword, reqPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(reqPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, nil // Indicate invalid password without revealing hashing details
		}
		return false, fmt.Errorf("failed to compare password: %w", err)
	}
	return true, nil // Password matches
}

func (s authService) createAccessToken(user param.LoginResponse) (string, error) {
	return s.createToken(user.User.ID, AccessTokenSubject, AccessTokenExpirationDuration)
}

func (s authService) refreshAccessToken(user param.LoginResponse) (string, error) {
	return s.createToken(user.User.ID, RefreshTokenSubject, RefreshTokenExpirationDuration)
}

// "github.com/golang-jwt/jwt/v4"
type Claims struct {
	jwt.RegisteredClaims
	UserID uint `json:"user_id"`
}

func (c Claims) Valid() error {
	return c.RegisteredClaims.Valid()
}

func (s authService) createToken(userID uint, subject string, expiresDuration time.Duration) (string, error) {
	// set our claims
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: subject,
			// set the expire time
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresDuration)),
		},
		UserID: userID,
	}

	// TODO add sign method to config
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := accessToken.SignedString([]byte(JwtSignKey))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s authService) VerifyToken(bearerToken string) (*Claims, error) {

	tokenStr := strings.Replace(bearerToken, "Bearer ", "", 1)

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JwtSignKey), nil
	})

	if err != nil {
		return nil, err
	}

	// convert interface to conceret object
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil

	} else {
		return nil, err
	}
}

// func (s authService) AddRevokedToken(tokenID string) error {
// 	panic("")
// }

// func (s authService) IsRevokedToken(tokenID string) error {
// 	panic("")
// }

// ! Middleware for token validation in Traefik
// func TokenValidationMiddleware(next http.Handler, authService authService) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Extract token from request
// 		token := ExtractTokenFromRequest(r)

// 		// Validate token
// 		if isValid := authService.ValidateToken(token); !isValid {
// 			// Token is invalid or revoked
// 			http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 			return
// 		}

// 		// Token is valid, proceed to next handler
// 		next.ServeHTTP(w, r)
// 	})
// }

// // ExtractTokenFromRequest extracts JWT token from request
// func ExtractTokenFromRequest(r *http.Request) string {
// 	// Extract token from request headers, cookies, or query parameters
// 	// Example: Authorization: Bearer <token>
// 	token := r.Header.Get("Authorization")
// 	if token != "" {
// 		return strings.TrimPrefix(token, "Bearer ")
// 	}

// 	// Extract token from cookies or query parameters if needed

// 	return ""
// }

// In a microservices architecture, it's generally a good practice to have each service be responsible for
// creating the queues it will consume from. This ensures that the service can function independently and
// that all necessary resources are in place when the service starts.
