package enums

type Permission string

const (
	CreateUser     Permission = "CREATE_USER"
	UpdateUser     Permission = "UPDATE_USER"
	DeleteUser     Permission = "DELETE_USER"
	ViewAllUsers   Permission = "VIEW_ALL_USERS"
	ManageRoles    Permission = "MANAGE_ROLES"
	ViewReports    Permission = "VIEW_REPORTS"
	ManageSettings Permission = "MANAGE_SETTINGS"
)
