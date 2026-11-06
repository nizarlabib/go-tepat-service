// TODO: DELETE THIS FILE

package handling

import (
	"errors"
)

func UnprocessableContent() error {
	return errors.New("422")
}

func BadRequest() error {
	return errors.New("400")
}

func Unauthorized() error {
	return errors.New("401")
}

func Forbidden() error {
	return errors.New("403")
}

func NotFound() error {
	return errors.New("404") // tidak ada
}

func Conflict() error {
	return errors.New("409") // duplicate
}

func InternalServerError() error {
	return errors.New("500")
}
