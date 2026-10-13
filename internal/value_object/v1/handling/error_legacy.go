// TODO: DELETE THIS FILE

package handling

import "net/http"

// UnauthorizedError struct jika logic buatkan baru error, tapi kalau logic di function, return err
type UnauthorizedError struct{}

func (e UnauthorizedError) Error() string {
	return http.StatusText(http.StatusUnauthorized)
}

type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	if e.Message == "" {
		return "Resource not found 404"
	}
	return e.Message
}
