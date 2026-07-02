package usecase

import (
	"github.com/theretech/retech-auth-api/internal/domain/repository"
)

// masterPermCode é o code emitido para usuários master — as APIs de recurso
// tratam como acesso total (mesma semântica do "manage all" do CASL).
const masterPermCode = "all:manage"

// buildPermCodes monta os codes de permission ("subject:action") que vão no
// claim `perms` do access token. Master colapsa para ["all:manage"] — token
// pequeno e semântica idêntica à do /me.
func buildPermCodes(roleCodes []string, permissions []*repository.PermissionInfo) []string {
	for _, code := range roleCodes {
		if code == "master" {
			return []string{masterPermCode}
		}
	}

	seen := make(map[string]struct{}, len(permissions))
	codes := make([]string, 0, len(permissions))
	for _, info := range permissions {
		if info == nil || info.Permission == nil {
			continue
		}
		p := info.Permission
		code := p.Code
		if code == "" && p.Subject != "" && p.Action != "" {
			code = p.Subject + ":" + p.Action
		}
		if code == "" {
			continue
		}
		if _, dup := seen[code]; dup {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	return codes
}
