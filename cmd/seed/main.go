package main

import (
	"context"
	"log"

	"github.com/theretech/retechauth-api/internal/application/service"
	"github.com/theretech/retechauth-api/internal/config"
	"github.com/theretech/retechauth-api/internal/domain/entity"
	"github.com/theretech/retechauth-api/internal/infrastructure/database"
	"github.com/theretech/retechauth-api/internal/infrastructure/repository"
)

func main() {
	// Carrega as configurações
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Erro ao carregar configurações: %v", err)
	}

	// Conecta ao banco de dados
	db, err := database.NewPostgresConnection(cfg.GetDSN())
	if err != nil {
		log.Fatalf("Erro ao conectar ao banco de dados: %v", err)
	}
	defer db.Close()

	log.Println("✅ Conectado ao banco de dados com sucesso!")

	// Inicializa repositories
	userRepo := repository.NewPostgresUserRepository(db)
	appRepo := repository.NewPostgresApplicationRepository(db)
	authRepo := repository.NewPostgresAuthRepository(db)
	hashService := service.NewHashService()

	ctx := context.Background()

	// 1. Aplicação de demonstração (alinhada ao ecossistema Retech)
	log.Println("📦 Criando aplicação retech-fin-admin...")
	demoApp := entity.NewApplication(
		"Retech Fin Admin",
		"retech-fin-admin",
		"Painel administrativo Retech Fin",
	)
	if err := appRepo.Create(ctx, demoApp); err != nil {
		log.Printf("⚠️  Aplicação já existe ou erro: %v", err)
	} else {
		log.Println("✅ Aplicação retech-fin-admin criada!")
	}

	// 2. Cria o usuário master
	log.Println("👤 Criando usuário master...")
	hashedPassword, err := hashService.HashPassword("Master@123")
	if err != nil {
		log.Fatalf("Erro ao gerar hash da senha: %v", err)
	}

	masterUser := entity.NewUser(
		"admin@theretech.local",
		hashedPassword,
		"Master Admin",
	)
	if err := userRepo.Create(ctx, masterUser); err != nil {
		log.Printf("⚠️  Usuário já existe ou erro: %v", err)
	} else {
		log.Println("✅ Usuário master criado!")
		log.Println("   📧 Email: admin@theretech.local")
		log.Println("   🔑 Senha: Master@123")
	}

	// 3. Vincula usuário à aplicação
	log.Println("🔗 Vinculando usuário à aplicação...")
	userApp := entity.NewUserApplication(masterUser.ID, demoApp.ID)
	if err := authRepo.CreateUserApplication(ctx, userApp); err != nil {
		log.Printf("⚠️  Vínculo já existe ou erro: %v", err)
	} else {
		log.Println("✅ Usuário vinculado à aplicação!")
	}

	// 4. Cria a role master
	log.Println("👑 Criando role master...")
	masterRole := entity.NewRole(
		demoApp.ID,
		"Master",
		"master",
		"Acesso total ao sistema",
	)
	if err := authRepo.CreateRole(ctx, masterRole); err != nil {
		log.Printf("⚠️  Role já existe ou erro: %v", err)
	} else {
		log.Println("✅ Role master criada!")
	}

	// 5. Atribui a role master ao usuário
	log.Println("🎭 Atribuindo role master ao usuário...")
	userRole := entity.NewUserRole(userApp.ID, masterRole.ID)
	if err := authRepo.AssignRoleToUser(ctx, userRole); err != nil {
		log.Printf("⚠️  Role já atribuída ou erro: %v", err)
	} else {
		log.Println("✅ Role master atribuída ao usuário!")
	}

	// 6. Permissões de exemplo para o painel administrativo
	log.Println("🔐 Criando permissões...")

	domainSubjects := []string{"User", "Role", "Application", "Permission", "Settings", "AuditLog"}
	menuSubjects := []string{"Menu:User", "Menu:Role", "Menu:Application", "Menu:Settings"}
	actions := []string{"manage"} // "manage" significa todas as ações (create, read, update, delete)

	var permissions []*entity.Permission
	for _, subject := range domainSubjects {
		for _, action := range actions {
			perm := entity.NewPermission(
				demoApp.ID,
				subject,
				action,
				"Gerenciar "+subject,
				nil,
			)
			permissions = append(permissions, perm)

			if err := authRepo.CreatePermission(ctx, perm); err != nil {
				log.Printf("⚠️  Permissão %s:%s já existe ou erro: %v", subject, action, err)
			} else {
				log.Printf("   ✅ Permissão %s:%s criada", subject, action)
			}
		}
	}

	for _, subject := range menuSubjects {
		perm := entity.NewPermission(
			demoApp.ID,
			subject,
			"view",
			"Visualizar "+subject,
			nil,
		)
		permissions = append(permissions, perm)

		if err := authRepo.CreatePermission(ctx, perm); err != nil {
			log.Printf("⚠️  Permissão %s:view já existe ou erro: %v", subject, err)
		} else {
			log.Printf("   ✅ Permissão %s:view criada", subject)
		}
	}

	// 7. Atribui todas as permissões à role master
	log.Println("🔗 Atribuindo permissões à role master...")
	for _, perm := range permissions {
		rolePerm := entity.NewRolePermission(masterRole.ID, perm.ID)
		if err := authRepo.AssignPermissionToRole(ctx, rolePerm); err != nil {
			log.Printf("⚠️  Permissão já atribuída ou erro: %v", err)
		}
	}
	log.Println("✅ Todas as permissões atribuídas à role master!")

	log.Println("\n🎉 Seed executado com sucesso!")
	log.Println("\n📝 Credenciais de teste:")
	log.Println("   Application Code: retech-fin-admin")
	log.Println("   Email: admin@theretech.local")
	log.Println("   Senha: Master@123")
	log.Println("\n💡 Use estas credenciais para testar a autenticação!")
}
