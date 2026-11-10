package permissionindex

const (
	SuperAdmin = iota

	Admin
	User

	ReadUser
	CreateUser
	UpdateUser
	DeleteUser

	ReadProduct
	CreateProduct
	UpdateProduct
	DeleteProduct
	UpdateProductStatus
)
