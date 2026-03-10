package domain

type EmailAlreadyInUseError struct {
	message string
}

func NewEmailAlreadyInUseError(message string) *EmailAlreadyInUseError {
	return &EmailAlreadyInUseError{message: message}
}
