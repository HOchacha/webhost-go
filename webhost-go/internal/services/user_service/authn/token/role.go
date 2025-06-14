package token

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleReader Role = "reader"
)

func isValidRole(role Role) bool {
	switch role {
	case RoleAdmin, RoleUser, RoleReader:
		return true
	default:
		return false
	}
}

func PassOnlyAdmin(claims *TokenValidationResult) bool {
	return claims.Valid && claims.Claims != nil && claims.Claims.Role == RoleAdmin
}

func PassOnlyUser(claims *TokenValidationResult) bool {
	return claims.Valid && claims.Claims != nil &&
		(claims.Claims.Role == RoleAdmin || claims.Claims.Role == RoleUser)
}

func PassOnlyReader(claims *TokenValidationResult) bool {
	return claims.Valid && claims.Claims != nil &&
		(claims.Claims.Role == RoleAdmin || claims.Claims.Role == RoleUser || claims.Claims.Role == RoleReader)
}
