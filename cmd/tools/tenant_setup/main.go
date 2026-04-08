package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type dbConfig struct {
	host     string
	port     string
	user     string
	password string
	dbName   string
	sslMode  string
}

type application struct {
	ID          uuid.UUID
	Name        string
	Code        string
	Description string
}

type user struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
}

type role struct {
	ID          uuid.UUID
	Name        string
	Code        string
	Description string
}

type permission struct {
	ID          uuid.UUID
	Subject     string
	Action      string
	Description string
	Conditions  *string
}

func main() {
	fmt.Println("==============================================")
	fmt.Println("🌱 Retech Auth - Assistente de Provisionamento")
	fmt.Println("==============================================")
	fmt.Println("Vamos criar uma tenant, um usuário, roles e permissões.")
	fmt.Println("Você pode usar tanto banco local quanto produção.")
	fmt.Println("")

	reader := bufio.NewReader(os.Stdin)

	cfg := promptDBConfig(reader)
	db := mustConnect(cfg)
	defer db.Close()

	fmt.Print("\n✅ Conexão com o banco realizada com sucesso!\n\n")

	app := ensureApplication(db, reader)
	usr := ensureUser(db, reader)
	userApplicationID := ensureUserApplication(db, usr.ID, app.ID)

	roles := provisionRolesAndPermissions(db, reader, app.ID)

	assignRolesToUser(db, reader, userApplicationID, roles)

	fmt.Println("\n==============================================")
	fmt.Println("🎉 Provisionamento concluído com sucesso!")
	fmt.Println("----------------------------------------------")
	fmt.Printf("Aplicação: %s (%s)\n", app.Name, app.Code)
	fmt.Printf("Usuário:   %s <%s>\n", usr.Name, usr.Email)
	fmt.Printf("Roles criadas: %d\n", len(roles))
	fmt.Println("----------------------------------------------")
	fmt.Println("Pronto! Já é possível autenticar no retech-auth-api usando esse usuário.")
	fmt.Println("Lembre-se de compartilhar as credenciais com segurança e atualizar a senha assim que possível.")
	fmt.Println("==============================================")
}

func promptDBConfig(reader *bufio.Reader) dbConfig {
	fmt.Println("Configuração do banco de dados:")
	fmt.Println("(Pressione Enter para usar os valores padrão)")

	host := readInput(reader, "Host", "localhost", false)
	port := readInput(reader, "Porta", "5432", false)
	user := readInput(reader, "Usuário", "retechauth", false)
	password := readInput(reader, "Senha", "retechauth_dev_password", false)
	dbName := readInput(reader, "Banco (dbname)", "retechauth_db", false)
	sslMode := readInput(reader, "SSL Mode (disable/require)", "disable", false)

	return dbConfig{
		host:     host,
		port:     port,
		user:     user,
		password: password,
		dbName:   dbName,
		sslMode:  sslMode,
	}
}

func mustConnect(cfg dbConfig) *sql.DB {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.host,
		cfg.port,
		cfg.user,
		cfg.password,
		cfg.dbName,
		cfg.sslMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Erro ao abrir conexão: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Erro ao conectar ao banco: %v", err)
	}

	return db
}

func ensureApplication(db *sql.DB, reader *bufio.Reader) application {
	fmt.Println("== Cadastro da Aplicação ==")
	for {
		code := strings.ToLower(readInput(reader, "Código da aplicação (slug)", "", true))
		name := readInput(reader, "Nome da aplicação", strings.Title(code), true)
		desc := readInput(reader, "Descrição", fmt.Sprintf("Aplicação %s", name), false)

		app, err := upsertApplication(db, name, code, desc)
		if err != nil {
			fmt.Printf("❌ Erro ao salvar aplicação: %v\n", err)
			continue
		}

		fmt.Printf("✅ Aplicação pronta! (%s - %s)\n", app.Code, app.Name)
		return app
	}
}

func upsertApplication(db *sql.DB, name, code, desc string) (application, error) {
	query := `
		INSERT INTO applications (id, name, code, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, NOW(), NOW())
		ON CONFLICT (code)
		DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, active = true, updated_at = NOW()
		RETURNING id, name, code, description
	`

	id := uuid.New()
	var app application
	err := db.QueryRow(query, id, name, code, desc).Scan(&app.ID, &app.Name, &app.Code, &app.Description)
	return app, err
}

func ensureUser(db *sql.DB, reader *bufio.Reader) user {
	fmt.Println("\n== Cadastro do Usuário ==")
	for {
		email := strings.ToLower(readInput(reader, "Email", "", true))
		usr, exists, err := findUserByEmail(db, email)
		if err != nil {
			fmt.Printf("❌ Erro ao buscar usuário: %v\n", err)
			continue
		}

		if exists {
			fmt.Printf("ℹ️  Usuário existente encontrado: %s (%s)\n", usr.Name, usr.Email)
			if confirm(reader, "Atualizar nome e senha?", false) {
				usr.Name = readInput(reader, "Nome", usr.Name, true)
				password := readInput(reader, "Nova senha", "", true)
				hash, err := hashPassword(password)
				if err != nil {
					fmt.Printf("❌ Erro ao gerar hash: %v\n", err)
					continue
				}

				if err := updateUser(db, usr.ID, usr.Name, hash); err != nil {
					fmt.Printf("❌ Erro ao atualizar usuário: %v\n", err)
					continue
				}
				fmt.Println("✅ Usuário atualizado!")
			}
			return usr
		}

		name := readInput(reader, "Nome", "", true)
		password := readInput(reader, "Senha", "", true)
		hash, err := hashPassword(password)
		if err != nil {
			fmt.Printf("❌ Erro ao gerar hash: %v\n", err)
			continue
		}

		usr, err = createUser(db, name, email, hash)
		if err != nil {
			fmt.Printf("❌ Erro ao criar usuário: %v\n", err)
			continue
		}

		fmt.Println("✅ Usuário criado!")
		return usr
	}
}

func findUserByEmail(db *sql.DB, email string) (user, bool, error) {
	query := `SELECT id, name, email FROM users WHERE email = $1`
	var usr user
	err := db.QueryRow(query, email).Scan(&usr.ID, &usr.Name, &usr.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return user{}, false, nil
		}
		return user{}, false, err
	}
	return usr, true, nil
}

func createUser(db *sql.DB, name, email, passwordHash string) (user, error) {
	query := `
		INSERT INTO users (id, email, password, name, active, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, true, 1, NOW(), NOW())
		RETURNING id, name, email
	`

	id := uuid.New()
	var usr user
	err := db.QueryRow(query, id, email, passwordHash, name).Scan(&usr.ID, &usr.Name, &usr.Email)
	if err != nil {
		return user{}, err
	}
	usr.Password = passwordHash
	return usr, nil
}

func updateUser(db *sql.DB, id uuid.UUID, name, passwordHash string) error {
	query := `
		UPDATE users SET name = $2, password = $3, active = true, updated_at = NOW()
		WHERE id = $1
	`
	_, err := db.Exec(query, id, name, passwordHash)
	return err
}

func ensureUserApplication(db *sql.DB, userID, appID uuid.UUID) uuid.UUID {
	query := `
		INSERT INTO user_applications (id, user_id, application_id, active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		ON CONFLICT (user_id, application_id)
		DO UPDATE SET active = true, updated_at = NOW()
		RETURNING id
	`

	id := uuid.New()
	var result uuid.UUID
	err := db.QueryRow(query, id, userID, appID).Scan(&result)
	if err != nil {
		log.Fatalf("Erro ao vincular usuário à aplicação: %v", err)
	}

	fmt.Println("✅ Usuário vinculado à aplicação!")
	return result
}

func provisionRolesAndPermissions(db *sql.DB, reader *bufio.Reader, appID uuid.UUID) []role {
	fmt.Println("\n== Roles & Permissões ==")

	var roles []role
	for {
		if !confirm(reader, "Deseja adicionar uma role?", len(roles) == 0) {
			break
		}

		code := strings.ToLower(readInput(reader, "Código da role (slug)", "", true))
		name := readInput(reader, "Nome da role", strings.Title(code), true)
		desc := readInput(reader, "Descrição da role", fmt.Sprintf("Role %s", name), false)

		role, err := upsertRole(db, appID, name, code, desc)
		if err != nil {
			fmt.Printf("❌ Erro ao salvar role: %v\n", err)
			continue
		}
		fmt.Printf("✅ Role pronta (%s)\n", role.Code)

		// Permissões
		permissions := provisionPermissions(db, reader, appID, role.ID)
		fmt.Printf("➡️  %d permissões vinculadas à role %s\n", len(permissions), role.Code)

		roles = append(roles, role)
	}

	if len(roles) == 0 {
		fmt.Println("⚠️  Nenhuma role criada. O usuário ficará sem permissões.")
	}

	return roles
}

func upsertRole(db *sql.DB, appID uuid.UUID, name, code, desc string) (role, error) {
	query := `
		INSERT INTO roles (id, application_id, name, code, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, true, NOW(), NOW())
		ON CONFLICT (application_id, code)
		DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description, active = true, updated_at = NOW()
		RETURNING id, name, code, description
	`

	id := uuid.New()
	var r role
	err := db.QueryRow(query, id, appID, name, code, desc).Scan(&r.ID, &r.Name, &r.Code, &r.Description)
	return r, err
}

func provisionPermissions(db *sql.DB, reader *bufio.Reader, appID, roleID uuid.UUID) []permission {
	var permissions []permission
	for {
		if !confirm(reader, "Adicionar permissão para esta role?", len(permissions) == 0) {
			break
		}

		fmt.Println("\nSugestões de subject:")
		fmt.Println("  • Menu:<Slug> (ex.: Menu:Dashboard)")
		fmt.Println("  • <Entidade> (ex.: Report, Device, User)")
		fmt.Println("  • Use o padrão que o CASL irá consumir no front")
		subject := readInput(reader, "Subject (recurso)", "", true)

		fmt.Println("Sugestões de action: manage, create, read, update, delete")
		action := readInput(reader, "Action", "manage", true)

		desc := readInput(reader, "Descrição da permissão", fmt.Sprintf("%s:%s", subject, action), false)
		condStr := readInput(reader, "Conditions (JSON opcional, Enter para nenhuma)", "", false)

		perm, err := upsertPermission(db, appID, subject, action, desc, condStr)
		if err != nil {
			fmt.Printf("❌ Erro ao salvar permissão: %v\n", err)
			continue
		}

		if err := attachPermissionToRole(db, roleID, perm.ID); err != nil {
			fmt.Printf("❌ Erro ao vincular permissão à role: %v\n", err)
			continue
		}

		fmt.Printf("   ✅ Permissão %s:%s pronta\n", perm.Subject, perm.Action)
		permissions = append(permissions, perm)
	}

	return permissions
}

func upsertPermission(db *sql.DB, appID uuid.UUID, subject, action, desc, conditions string) (permission, error) {
	query := `
		INSERT INTO permissions (id, application_id, subject, action, conditions, description, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, true, NOW(), NOW())
		ON CONFLICT (application_id, subject, action)
		DO UPDATE SET description = EXCLUDED.description, conditions = EXCLUDED.conditions, active = true, updated_at = NOW()
		RETURNING id, subject, action, description, conditions
	`

	id := uuid.New()
	var perm permission

	var condPtr interface{}
	if strings.TrimSpace(conditions) == "" {
		condPtr = nil
	} else {
		condPtr = conditions
	}

	err := db.QueryRow(query, id, appID, subject, action, condPtr, desc).Scan(&perm.ID, &perm.Subject, &perm.Action, &perm.Description, &perm.Conditions)
	return perm, err
}

func attachPermissionToRole(db *sql.DB, roleID, permissionID uuid.UUID) error {
	query := `
		INSERT INTO role_permissions (id, role_id, permission_id, active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		ON CONFLICT (role_id, permission_id)
		DO UPDATE SET active = true, updated_at = NOW()
	`

	_, err := db.Exec(query, uuid.New(), roleID, permissionID)
	return err
}

func assignRolesToUser(db *sql.DB, reader *bufio.Reader, userApplicationID uuid.UUID, roles []role) {
	if len(roles) == 0 {
		return
	}

	fmt.Println("\n== Vincular roles ao usuário ==")
	for _, r := range roles {
		if confirm(reader, fmt.Sprintf("Atribuir role %s ao usuário?", r.Code), true) {
			if err := attachRoleToUser(db, userApplicationID, r.ID); err != nil {
				fmt.Printf("❌ Erro ao atribuir role %s: %v\n", r.Code, err)
			} else {
				fmt.Printf("✅ Role %s atribuída!\n", r.Code)
			}
		}
	}
}

func attachRoleToUser(db *sql.DB, userApplicationID, roleID uuid.UUID) error {
	query := `
		INSERT INTO user_roles (id, user_application_id, role_id, active, created_at, updated_at)
		VALUES ($1, $2, $3, true, NOW(), NOW())
		ON CONFLICT (user_application_id, role_id)
		DO UPDATE SET active = true, updated_at = NOW()
	`

	_, err := db.Exec(query, uuid.New(), userApplicationID, roleID)
	return err
}

func readInput(reader *bufio.Reader, label, defaultValue string, required bool) string {
	for {
		prompt := label
		if defaultValue != "" {
			prompt = fmt.Sprintf("%s [%s]", label, defaultValue)
		}
		fmt.Printf("%s: ", prompt)
		value, _ := reader.ReadString('\n')
		value = strings.TrimSpace(value)

		if value == "" {
			if defaultValue != "" {
				return defaultValue
			}
			if !required {
				return ""
			}
			fmt.Println("⚠️  Este campo é obrigatório")
			continue
		}

		return value
	}
}

func confirm(reader *bufio.Reader, question string, defaultYes bool) bool {
	def := "N"
	if defaultYes {
		def = "S"
	}

	for {
		fmt.Printf("%s (s/n) [%s]: ", question, def)
		value, _ := reader.ReadString('\n')
		value = strings.ToLower(strings.TrimSpace(value))

		if value == "" {
			return defaultYes
		}

		if value == "s" || value == "sim" {
			return true
		}
		if value == "n" || value == "nao" || value == "não" {
			return false
		}

		fmt.Println("⚠️  Responda com s (sim) ou n (não)")
	}
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// Utilitário para logs com timestamps
func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
