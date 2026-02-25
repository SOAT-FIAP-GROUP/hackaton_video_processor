package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"frontend/internal/domain"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
)

type CognitoAuth struct {
	client       *cognitoidentityprovider.Client
	clientID     string
	clientSecret string
	userPoolID   string
}

func NewCognitoClient(region, clientID, clientSecret, userPoolId string) (*CognitoAuth, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &CognitoAuth{
		client:       cognitoidentityprovider.NewFromConfig(cfg),
		clientID:     clientID,
		clientSecret: clientSecret,
		userPoolID:   userPoolId,
	}, nil
}

func (c *CognitoAuth) computeSecretHash(clientSecret, username, clientID string) string {
	mac := hmac.New(sha256.New, []byte(clientSecret))
	mac.Write([]byte(username + clientID))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func (c *CognitoAuth) SignUp(email, password, name string) (*cognitoidentityprovider.SignUpOutput, error) {
	secretHash := c.computeSecretHash(c.clientSecret, email, c.clientID)

	input := &cognitoidentityprovider.SignUpInput{
		ClientId:   aws.String(c.clientID),
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
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(email),
	})
	if err != nil {
		log.Printf("Admin confirm failed: %v", err)
	} else {
		log.Printf("Admin confirm successful: %v", c)
	}

	return result, nil
}

func (c *CognitoAuth) Login(email, password string) (*cognitoidentityprovider.InitiateAuthOutput, error) {
	secretHash := c.computeSecretHash(c.clientSecret, email, c.clientID)

	input := &cognitoidentityprovider.InitiateAuthInput{
		ClientId: aws.String(c.clientID),
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

func (c *CognitoAuth) AdminConfirmUser(email string) error {
	input := &cognitoidentityprovider.AdminConfirmSignUpInput{
		UserPoolId: aws.String(c.userPoolID),
		Username:   aws.String(email),
	}

	_, err := c.client.AdminConfirmSignUp(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("admin confirm failed: %w", err)
	}

	return nil
}

func (c *CognitoAuth) ParseToken(tokenString string) (*domain.Claims, error) {
	result, err := c.client.GetUser(context.Background(), &cognitoidentityprovider.GetUserInput{
		AccessToken: &tokenString,
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims := &domain.Claims{}
	for _, attr := range result.UserAttributes {
		switch *attr.Name {
		case "sub":
			claims.UserID = *attr.Value
		case "email":
			claims.Email = *attr.Value
		}
	}

	return claims, nil
}
