package main

import (
	"context"
	"fmt"
	"log"

	"github.com/theretech/retechauth-api/internal/application/service"
	"github.com/theretech/retechauth-api/internal/config"
	"github.com/theretech/retechauth-api/internal/infrastructure/database"
	"github.com/google/uuid"
)

func main() {
	log.Println("🚀 Configurando usuário Master...")

	// Carrega as configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Erro ao carregar configurações: %v", err)
	}

	// Conecta ao banco de dados
	db, err := database.NewPostgresConnection(cfg.GetDSN())
	if err != nil {
		log.Fatalf("❌ Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	log.Println("✅ Conectado ao banco de dados")

	ctx := context.Background()

	// IDs fixos para facilitar referências
	appID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	roleID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	userID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
	userAppID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")
	userRoleID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440005")

	// 1. Criar aplicação retech-fin-admin
	log.Println("📦 Criando aplicação retech-fin-admin...")
	_, err = db.ExecContext(ctx, `
		INSERT INTO applications (id, name, code, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at = NOW()
	`, appID, "Retech Fin Admin", "retech-fin-admin", "Painel administrativo Retech Fin")
	if err != nil {
		log.Fatalf("❌ Erro ao criar aplicação: %v", err)
	}
	log.Println("✅ Aplicação retech-fin-admin criada/atualizada")

	// 2. Criar role Master
	log.Println("👑 Criando role Master...")
	_, err = db.ExecContext(ctx, `
		INSERT INTO roles (id, application_id, name, code, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
		ON CONFLICT (application_id, code) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			updated_at = NOW()
	`, roleID, appID, "Master", "master", "Acesso total ao sistema")
	if err != nil {
		log.Fatalf("❌ Erro ao criar role: %v", err)
	}
	log.Println("✅ Role Master criada/atualizada")

	// 3. Gerar hash da senha (mesmo padrão do seed)
	log.Println("🔐 Gerando hash da senha...")
	hashService := service.NewHashService()
	passwordHash, err := hashService.HashPassword("Master@123")
	if err != nil {
		log.Fatalf("❌ Erro ao gerar hash da senha: %v", err)
	}

	// 4. Criar usuário Master
	log.Println("👤 Criando usuário Master...")
	_, err = db.ExecContext(ctx, `
		INSERT INTO users (id, email, password, name, active, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, 1, NOW(), NOW())
		ON CONFLICT (email) DO UPDATE SET
			password = EXCLUDED.password,
			name = EXCLUDED.name,
			active = EXCLUDED.active,
			updated_at = NOW()
	`, userID, "admin@theretech.local", passwordHash, "Master Admin")
	if err != nil {
		log.Fatalf("❌ Erro ao criar usuário: %v", err)
	}
	log.Println("✅ Usuário admin@theretech.local criado/atualizado")

	// 5. Vincular usuário à aplicação
	log.Println("🔗 Vinculando usuário à aplicação...")
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_applications (id, user_id, application_id, active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		ON CONFLICT (user_id, application_id) DO UPDATE SET
			active = EXCLUDED.active,
			updated_at = NOW()
	`, userAppID, userID, appID)
	if err != nil {
		log.Fatalf("❌ Erro ao vincular usuário à aplicação: %v", err)
	}
	log.Println("✅ Usuário vinculado à aplicação retech-fin-admin")

	// 6. Vincular usuário à role Master
	log.Println("👑 Atribuindo role Master ao usuário...")
	_, err = db.ExecContext(ctx, `
		INSERT INTO user_roles (id, user_application_id, role_id, active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		ON CONFLICT (user_application_id, role_id) DO UPDATE SET
			active = EXCLUDED.active,
			updated_at = NOW()
	`, userRoleID, userAppID, roleID)
	if err != nil {
		log.Fatalf("❌ Erro ao atribuir role: %v", err)
	}
	log.Println("✅ Role Master atribuída")

	// Verificação final
	log.Println("\n🔍 Verificando configuração...")
	var email, name, appCode, roleCode, roleName string
	err = db.QueryRowContext(ctx, `
		SELECT 
			u.email,
			u.name,
			a.code AS application,
			r.code AS role,
			r.name AS role_name
		FROM users u
		JOIN user_applications ua ON u.id = ua.user_id
		JOIN applications a ON ua.application_id = a.id
		JOIN user_roles ur ON ua.id = ur.user_application_id
		JOIN roles r ON ur.role_id = r.id
		WHERE u.email = $1
		AND a.code = $2
	`, "admin@theretech.local", "retech-fin-admin").Scan(&email, &name, &appCode, &roleCode, &roleName)

	if err != nil {
		log.Fatalf("❌ Erro na verificação: %v", err)
	}

	fmt.Println("\n✅ ✨ Configuração concluída com sucesso! ✨")
	fmt.Println("\n📋 Detalhes do usuário Master:")
	fmt.Printf("   Email:       %s\n", email)
	fmt.Printf("   Nome:        %s\n", name)
	fmt.Printf("   Senha:       Master@123\n")
	fmt.Printf("   Aplicação:   %s\n", appCode)
	fmt.Printf("   Role:        %s (%s)\n", roleName, roleCode)
	baseURL := cfg.Docs.APIBaseURL

	fmt.Println("\n🧪 Para testar, execute:")
	fmt.Printf("\n   curl -X POST %s/authenticate \\\n", baseURL)
	fmt.Println("     -H \"Content-Type: application/json\" \\")
	fmt.Println("     -d '{")
	fmt.Println("       \"email\": \"admin@theretech.local\",")
	fmt.Println("       \"password\": \"Master@123\",")
	fmt.Println("       \"application_code\": \"retech-fin-admin\"")
	fmt.Println("     }'")
	fmt.Println("\n   # Depois:")
	fmt.Printf("   curl -X GET %s/me \\\n", baseURL)
	fmt.Println("     -H \"Authorization: Bearer <seu_token>\"")
	fmt.Println()
}
