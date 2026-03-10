package domain

type AuthServiceInterface interface {
	GenerateJWT(userID, email string) (string, error)
	ParseToken(tokenString string) (*Claims, error)
}
