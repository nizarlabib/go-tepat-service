package permissionindex

type Permission struct {
	Bit  int
	Name string
}

var Permissions = []*Permission{
	{SuperAdmin, "Super Admin"},

	{ReadUser, "Read User"},
	{CreateUser, "Create User"},
	{UpdateUser, "Update User"},
	{DeleteUser, "Delete User"},

	{ReadProduct, "Read Product"},
	{CreateProduct, "Create Product"},
	{UpdateProduct, "Update Product"},
	{DeleteProduct, "Delete Product"},
	{UpdateProductStatus, "Update Product Status"},
}
