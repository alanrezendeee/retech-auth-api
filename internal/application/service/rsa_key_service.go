package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/crypto/sha3"
)

var (
	ErrKeyNotFound = errors.New("chave RSA não encontrada")
	ErrInvalidKey  = errors.New("chave RSA inválida")
)

// RSAKeyService gerencia chaves RSA para assinatura de tokens JWT
type RSAKeyService interface {
	GetPrivateKey(kid string) (*rsa.PrivateKey, error)
	GetPublicKey(kid string) (*rsa.PublicKey, error)
	GetCurrentKeyID() string
	GetPublicKeyJWK(kid string) (map[string]interface{}, error)
	GetAllPublicKeysJWK() ([]map[string]interface{}, error)
	RotateKey() (string, error)
}

type rsaKeyService struct {
	keysDir       string
	currentKeyID  string
	keys          map[string]*rsa.PrivateKey
	mu            sync.RWMutex
	keySize       int
}

// NewRSAKeyService cria uma nova instância de RSAKeyService
// Se keysDir for vazio, usa diretório padrão "./keys"
func NewRSAKeyService(keysDir string) (RSAKeyService, error) {
	if keysDir == "" {
		keysDir = "./keys"
	}

	// Cria diretório se não existir
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de chaves: %w", err)
	}

	service := &rsaKeyService{
		keysDir: keysDir,
		keys:    make(map[string]*rsa.PrivateKey),
		keySize: 2048, // Tamanho padrão de chave RSA
	}

	// Carrega chaves existentes ou gera uma nova
	if err := service.loadOrGenerateKeys(); err != nil {
		return nil, fmt.Errorf("erro ao carregar/gerar chaves: %w", err)
	}

	return service, nil
}

// loadOrGenerateKeys carrega chaves existentes do disco ou gera uma nova
func (s *rsaKeyService) loadOrGenerateKeys() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Lista arquivos .pem no diretório
	files, err := filepath.Glob(filepath.Join(s.keysDir, "*.pem"))
	if err != nil {
		return err
	}

	// Carrega todas as chaves encontradas
	for _, file := range files {
		kid := filepath.Base(file[:len(file)-4]) // Remove .pem
		privateKey, err := s.loadPrivateKeyFromFile(file)
		if err != nil {
			continue // Ignora arquivos inválidos
		}
		s.keys[kid] = privateKey
		if s.currentKeyID == "" || kid > s.currentKeyID {
			s.currentKeyID = kid // Usa a chave mais recente (lexicograficamente maior)
		}
	}

	// Se não encontrou nenhuma chave, gera uma nova
	if s.currentKeyID == "" {
		kid, err := s.generateNewKey()
		if err != nil {
			return err
		}
		s.currentKeyID = kid
	}

	return nil
}

// generateNewKey gera um novo par de chaves RSA e salva no disco
func (s *rsaKeyService) generateNewKey() (string, error) {
	// Gera chave privada
	privateKey, err := rsa.GenerateKey(rand.Reader, s.keySize)
	if err != nil {
		return "", fmt.Errorf("erro ao gerar chave RSA: %w", err)
	}

	// Gera Key ID baseado no hash da chave pública
	kid := s.generateKeyID(privateKey.PublicKey)

	// Salva chave privada no disco
	if err := s.savePrivateKeyToFile(kid, privateKey); err != nil {
		return "", err
	}

	// Armazena em memória
	s.keys[kid] = privateKey

	return kid, nil
}

// generateKeyID gera um Key ID único baseado na chave pública
func (s *rsaKeyService) generateKeyID(publicKey rsa.PublicKey) string {
	// Serializa a chave pública
	pubKeyBytes := x509.MarshalPKCS1PublicKey(&publicKey)
	// Gera hash SHA3-256
	hash := sha3.Sum256(pubKeyBytes)
	// Retorna primeiros 8 bytes em base64url (kid curto e único)
	return base64.RawURLEncoding.EncodeToString(hash[:8])
}

// savePrivateKeyToFile salva a chave privada em arquivo PEM
func (s *rsaKeyService) savePrivateKeyToFile(kid string, privateKey *rsa.PrivateKey) error {
	filePath := filepath.Join(s.keysDir, kid+".pem")
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo de chave: %w", err)
	}
	defer file.Close()

	// Converte para PEM
	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	if err := pem.Encode(file, block); err != nil {
		return fmt.Errorf("erro ao codificar chave PEM: %w", err)
	}

	return nil
}

// loadPrivateKeyFromFile carrega a chave privada de um arquivo PEM
func (s *rsaKeyService) loadPrivateKeyFromFile(filePath string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrInvalidKey
	}

	if block.Type != "RSA PRIVATE KEY" {
		return nil, ErrInvalidKey
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

// GetPrivateKey retorna a chave privada para um Key ID
func (s *rsaKeyService) GetPrivateKey(kid string) (*rsa.PrivateKey, error) {
	if kid == "" {
		kid = s.GetCurrentKeyID()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	privateKey, ok := s.keys[kid]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return privateKey, nil
}

// GetPublicKey retorna a chave pública para um Key ID
func (s *rsaKeyService) GetPublicKey(kid string) (*rsa.PublicKey, error) {
	if kid == "" {
		kid = s.GetCurrentKeyID()
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	privateKey, ok := s.keys[kid]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return &privateKey.PublicKey, nil
}

// GetCurrentKeyID retorna o ID da chave atual
func (s *rsaKeyService) GetCurrentKeyID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentKeyID
}

// GetPublicKeyJWK retorna a chave pública no formato JWK (JSON Web Key)
func (s *rsaKeyService) GetPublicKeyJWK(kid string) (map[string]interface{}, error) {
	if kid == "" {
		kid = s.GetCurrentKeyID()
	}

	publicKey, err := s.GetPublicKey(kid)
	if err != nil {
		return nil, err
	}

	// Extrai modulus (n) e exponent (e) da chave pública RSA
	// N é *big.Int, E é int
	nBytes := publicKey.N.Bytes()
	
	// Converte E (int) para big.Int e depois para bytes
	// Isso garante que valores maiores sejam tratados corretamente
	eBig := big.NewInt(int64(publicKey.E))
	eBytes := eBig.Bytes()

	// Encodifica em base64url (sem padding)
	n := base64.RawURLEncoding.EncodeToString(nBytes)
	e := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := map[string]interface{}{
		"kty": "RSA",
		"kid": kid,
		"use": "sig",
		"alg": "RS256",
		"n":   n,
		"e":   e,
	}

	return jwk, nil
}

// GetAllPublicKeysJWK retorna todas as chaves públicas no formato JWK
func (s *rsaKeyService) GetAllPublicKeysJWK() ([]map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var jwks []map[string]interface{}
	for kid := range s.keys {
		jwk, err := s.GetPublicKeyJWK(kid)
		if err != nil {
			continue // Ignora erros em chaves individuais
		}
		jwks = append(jwks, jwk)
	}

	return jwks, nil
}

// RotateKey gera uma nova chave e a define como atual
// Mantém as chaves antigas para permitir validação de tokens antigos durante período de transição
func (s *rsaKeyService) RotateKey() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Gera nova chave
	newKid, err := s.generateNewKey()
	if err != nil {
		return "", err
	}

	// Define como chave atual
	oldKid := s.currentKeyID
	s.currentKeyID = newKid

	// Mantém chaves antigas em memória (podem ser removidas manualmente após período de transição)
	_ = oldKid // Chaves antigas permanecem no map

	return newKid, nil
}

