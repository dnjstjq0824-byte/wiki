package auth

import (
	"time"
)

const (
	RoleAnonymous    = "ANONYMOUS"
	RoleRegistered   = "REGISTERED"
	RoleAutoConfirmed = "AUTOCONFIRMED"
	RoleRollbacker   = "ROLLBACKER"
	RoleEditor       = "EDITOR"
	RoleSysop        = "SYSOP"
	RoleOversight    = "OVERSIGHT"
	RoleBureaucrat   = "BUREAUCRAT"
	RoleSystemAdmin  = "SYSTEM_ADMIN"
)

var roleLevels = map[string]int{
	RoleAnonymous:     0,
	RoleRegistered:    10,
	RoleAutoConfirmed: 20,
	RoleRollbacker:    30,
	RoleEditor:        40,
	RoleSysop:         50,
	RoleOversight:     60,
	RoleBureaucrat:    70,
	RoleSystemAdmin:   100,
}

// RoleLevel returns the numeric hierarchy level for an RBAC role
func RoleLevel(role string) int {
	if lvl, ok := roleLevels[role]; ok {
		return lvl
	}
	return 0
}

// HasPermission checks if the user's role meets or exceeds the required threshold
func HasPermission(userRole, requiredRole string) bool {
	return RoleLevel(userRole) >= RoleLevel(requiredRole)
}

// CheckAutoConfirmed evaluates tenure and minimum contribution requirements
// Standard policy: Account registered >= 7 days ago and edit count >= 15
func CheckAutoConfirmed(registeredAt time.Time, editCount int) bool {
	tenure := time.Since(registeredAt)
	return tenure >= 7*24*time.Hour && editCount >= 15
}

// EffectiveRole calculates the actual dynamic role (e.g. promoting REGISTERED to AUTOCONFIRMED)
func EffectiveRole(baseRole string, registeredAt time.Time, editCount int) string {
	if baseRole == RoleRegistered && CheckAutoConfirmed(registeredAt, editCount) {
		return RoleAutoConfirmed
	}
	return baseRole
}

