package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"log"

	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/gin-gonic/gin"
)

const (
	clientID     = "7ce79qdsoue6scgt8rutqkg2g5"
	clientSecret = "191qjpnr4cdevilek4jll0l170vjg4lcdbgvecjkd99st7emjrdu"
	userPoolID   = "us-east-1_9FLy5Ac7e"
)

type CognitoClient struct {
	client *cognitoidentityprovider.Client
}

func NewCognitoClient(region string) (*CognitoClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &CognitoClient{
		client: cognitoidentityprovider.NewFromConfig(cfg),
	}, nil
}

func computeSecretHash(clientSecret, username, clientID string) string {
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(username + clientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (c *CognitoClient) SignUp(email, password, name string) (*cognitoidentityprovider.SignUpOutput, error) {
	secretHash := computeSecretHash(clientSecret, email, clientID)

	input := &cognitoidentityprovider.SignUpInput{
		ClientId:   aws.String(clientID),
		Username:   aws.String(email),
		Password:   aws.String(password),
		SecretHash: aws.String(secretHash),
		UserAttributes: []types.AttributeType{
			{Name: aws.String("email"), Value: aws.String(email)},
			{Name: aws.String("name"), Value: aws.String(name)},
		},
	}

	result, err := c.client.SignUp(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("sign up failed: %w", err)
	}

	_, err = c.client.AdminConfirmSignUp(context.TODO(), &cognitoidentityprovider.AdminConfirmSignUpInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(email),
	})
	if err != nil {
		log.Printf("Admin confirm failed: %v", err)
	} else {
		log.Printf("Admin confirm successful: %v", c)
	}

	return result, nil
}

func (c *CognitoClient) Login(email, password string) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	secretHash := computeSecretHash(clientSecret, email, clientID)

	input := &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(clientID),
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		AuthParameters: map[string]string{
			"USERNAME":    email,
			"PASSWORD":    password,
			"SECRET_HASH": secretHash,
		},
	}

	result, err := c.client.InitiateAuth(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	return result, nil
}

func (c *CognitoClient) AdminConfirmUser(email string) error {
	input := &cognitoidentityprovider.AdminConfirmSignUpInput{
		UserPoolId: aws.String(userPoolID),
		Username:   aws.String(email),
	}

	_, err := c.client.AdminConfirmSignUp(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("admin confirm failed: %w", err)
	}

	return nil
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"     binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func LoginHandler(cognitoClient *CognitoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := cognitoClient.Login(req.Email, req.Password)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"accessToken":  result.AuthenticationResult.AccessToken,
			"idToken":      result.AuthenticationResult.IdToken,
			"refreshToken": result.AuthenticationResult.RefreshToken,
			"expiresIn":    result.AuthenticationResult.ExpiresIn,
		})
	}
}

func RegisterHandler(cognitoClient *CognitoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := cognitoClient.SignUp(req.Email, req.Password, req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message":   "User registered successfully",
			"userSub":   result.UserSub,
			"confirmed": result.UserConfirmed,
		})
	}
}

func main() {
	cognitoClient, err := NewCognitoClient("us-east-1")
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	r.POST("/register", RegisterHandler(cognitoClient))
	r.POST("/login", LoginHandler(cognitoClient))
	r.Run(":8080")
}
