package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type manager struct {
	config Config
}

func NewManager(config Config) Service {
	return &manager{
		config: config,
	}
}

func (m *manager) HashPassword(password string) (string, error) {
	if err := m.ValidatePassword(password); err != nil {
		return "", err
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), m.config.Cost)
	return string(bytes), err
}

func (m *manager) VerifyPassword(hashedPassword, password string) error {
	if err := m.ValidatePassword(password); err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (m *manager) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	if len(password) > 128 {
		return errors.New("password must not exceed 128 characters")
	}

	return nil
}
