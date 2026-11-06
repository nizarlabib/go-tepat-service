package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/id"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"

	en_trans "github.com/go-playground/validator/v10/translations/en"
	id_trans "github.com/go-playground/validator/v10/translations/id"
)

var myValidator *MyValidator

func init() {
	myValidator = NewMyValidator()
}

type MyValidator struct {
	Validate *validator.Validate
	UniTrans *ut.UniversalTranslator

	Trans ut.Translator
}

func NewMyValidator() *MyValidator {
	myV := &MyValidator{
		Validate: newValidator(),
	}
	registerTranslations(myV)
	registerCustomValidators(myV)
	return myV
}

func newValidator() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return v
}

func registerTranslations(myV *MyValidator) {
	v := myV.Validate

	id := id.New()
	en := en.New()

	fl := en

	uniTrans := ut.New(fl, id, en)

	transID, _ := uniTrans.GetTranslator("id")
	transEN, _ := uniTrans.GetTranslator("en")

	_ = id_trans.RegisterDefaultTranslations(v, transID)
	_ = en_trans.RegisterDefaultTranslations(v, transEN)

	myV.UniTrans = uniTrans
	myV.Trans = transID
}
