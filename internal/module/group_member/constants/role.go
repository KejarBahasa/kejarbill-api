package constants

const (
	RoleMember = "member"

	RoleAdmin = "admin"

	RoleOwner = "owner"
)

func CanManage(role string) bool {
	return role == RoleOwner || role == RoleAdmin
}
