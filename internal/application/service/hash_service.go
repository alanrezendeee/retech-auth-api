package service

import "golang.org/x/crypto/bcrypt"

// HashService fornece métodos para hash e verificação de senhas
type HashService interface {
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword, password string) error
}

type bcryptHashService struct {
	cost int
}

// NewHashService cria uma nova instância de HashService
func NewHashService() HashService {
	return &bcryptHashService{
		cost: bcrypt.DefaultCost,
	}
}

// HashPassword gera um hash bcrypt da senha
func (s *bcryptHashService) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword verifica se a senha corresponde ao hash
func (s *bcryptHashService) CheckPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

