package auth

type Role string

const (
	RoleCustomer Role = "customer"
	RoleManager  Role = "manager"
	RoleCourier  Role = "courier"
	RoleAdmin    Role = "admin"
)

func (r Role) String() string {
	return string(r)
}
