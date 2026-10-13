package validator

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func registerCustomValidators(myV *MyValidator) {
	v := myV.Validate

	v.RegisterValidation("phone", validatePhone, false)
	v.RegisterValidation("password", validatePassword, false)
	v.RegisterValidation("date", validateDate, false) // Format Indonesia DD-MM-YYYY
	v.RegisterValidation("filename", validateFilename, false)
	v.RegisterValidation("optional", validateOptional, true) // backward compatibility
	v.RegisterValidation("required_if_top", validateRequiredIfTop, true)

	transID, _ := myV.UniTrans.GetTranslator("id")
	_ = v.RegisterTranslation("phone", transID, func(ut ut.Translator) error {
		return ut.Add("phone", "{0} harus menggunakan format nomor telepon yang valid.", false)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T(fe.Tag(), fe.Field())
		if err != nil {
			log.Printf("warning: error translating FieldError: %#v", fe)
			return fe.(error).Error()
		}
		return t
	})
	_ = v.RegisterTranslation("date", transID, func(ut ut.Translator) error {
		return ut.Add("date", "{0} harus menggunakan format tanggal DD-MM-YYYY.", false)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T(fe.Tag(), fe.Field())
		if err != nil {
			log.Printf("warning: error translating FieldError: %#v", fe)
			return fe.(error).Error()
		}
		return t
	})
	v.RegisterTranslation("required_if_top", transID, func(ut ut.Translator) error {
		return ut.Add("required_if_top", "{0} wajib diisi jika {1}", false) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T(fe.Tag(), fe.Field(), fe.Param())
		if err != nil {
			log.Printf("warning: error translating FieldError: %#v", fe)
			return fe.(error).Error()
		}

		return t
	})
	v.RegisterTranslation("password", transID, func(ut ut.Translator) error {
		return ut.Add("password", "{0} anda lemah atau tidak aman", false) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, err := ut.T(fe.Tag(), fe.Field(), fe.Param())
		if err != nil {
			log.Printf("warning: error translating FieldError: %#v", fe)
			return fe.(error).Error()
		}

		return t
	})
}

// validatePhone memvalidasi format nomor telepon
func validatePhone(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	if phone == "" {
		return true
	}

	const phoneRegex = `\b[+()]?(?<country>1939|1876|1869|1868|1849|1784|1767|1758|1684|1671|1670|1664|1649|1473|1441|1345|1340|1284|1268|1264|1246|1242|998|996|995|994|993|992|977|976|975|974|973|972|971|970|968|967|966|965|964|963|962|961|960|886|880|856|855|853|852|850|692|691|690|689|688|687|686|685|683|682|681|680|679|678|677|676|675|674|673|672|670|598|597|596|595|594|593|592|591|590|509|508|507|506|505|504|503|502|501|500|423|421|420|389|387|386|385|382|381|380|379|378|377|376|375|374|373|372|371|370|359|358|357|356|355|354|353|352|351|350|299|298|297|291|290|269|268|267|266|265|264|263|262|261|260|258|257|256|255|254|253|252|251|250|249|248|246|245|244|243|242|241|240|239|238|237|236|235|234|233|232|231|230|229|228|227|226|225|224|223|222|221|220|218|216|213|212|211|98|95|94|93|92|91|90|86|84|82|81|66|65|64|63|62|61|60|58|57|56|55|54|53|52|51|49|48|47|46|45|44|43|41|40|39|36|34|33|32|31|30|27|20|7|1)[+()]?(?<number>[\d]{4,})\b`
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if password == "" {
		return true
	}

	level := PasswordStrengthMeter(password)
	return level != Terrible
}

func validateDate(fl validator.FieldLevel) bool {
	dateStr := fl.Field().String()
	if dateStr == "" {
		return true
	}

	_, err := time.Parse("02-01-2006", dateStr)
	return err == nil
}

func validateFilename(fl validator.FieldLevel) bool {
	filename := fl.Field().String()
	if filename == "" {
		return true
	}

	err := IsValidFilename(filename)
	return err == nil
}

// custom validator 'optional' for backward compatibility
func validateOptional(fl validator.FieldLevel) bool {
	// always return true
	return true
}

var (
	paramValsCache  = map[string][]string{}
	paramValsRWLock = sync.RWMutex{}
)

func parseParams(s string) []string {
	paramValsRWLock.RLock()
	vals, found := paramValsCache[s]
	paramValsRWLock.RUnlock()
	if !found {
		paramValsRWLock.Lock()
		vals = splitParamsRegex().FindAllString(s, -1)
		for i := range vals {
			vals[i] = strings.ReplaceAll(vals[i], "'", "")
		}
		paramValsCache[s] = vals
		paramValsRWLock.Unlock()
	}
	return vals
}

func validateRequiredIfTop(fl validator.FieldLevel) bool {
	params := parseParams(fl.Param()) // [Status, submit, Type, sk]

	if len(params)%2 != 0 {
		panic(fmt.Sprintf("Bad param number for required_if %s", fl.FieldName()))
	}

	seen := make(map[string]struct{})
	for i := 0; i < len(params); i += 2 {
		if _, ok := seen[params[i]]; ok {
			panic(fmt.Sprintf("Duplicate param %s for required_if %s", params[i], fl.FieldName()))
		}
		seen[params[i]] = struct{}{}
	}

	for i := 0; i < len(params); i += 2 {
		if !requireCheckFieldValueFromTop(fl, params[i], params[i+1], false) {
			return true
		}
	}

	return hasValue(fl)
}

func hasValue(fl validator.FieldLevel) bool {
	field := fl.Field()

	switch field.Kind() {
	case reflect.Slice, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Chan, reflect.Func:
		return !field.IsNil()
	case reflect.String:
		return strings.TrimSpace(field.String()) != ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return field.Float() != 0.0
	case reflect.Bool:
		return true
	default:
		return field.IsValid() && !field.IsZero()
	}
}

func requireCheckFieldValueFromTop(
	fl validator.FieldLevel, param string, value string, defaultNotFoundValue bool,
) bool {
	field, kind, _, found := fl.GetStructFieldOKAdvanced2(fl.Top(), param)
	if !found {
		return defaultNotFoundValue
	}

	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == asInt(value)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint() == asUint(value)

	case reflect.Float32:
		return field.Float() == asFloat32(value)

	case reflect.Float64:
		return field.Float() == asFloat64(value)

	case reflect.Slice, reflect.Map:
		if value == "nil" {
			return field.IsNil()
		}
		return int64(field.Len()) == asInt(value)
	case reflect.Array:
		// Arrays can't be nil, so only compare lengths
		return int64(field.Len()) == asInt(value)

	case reflect.Bool:
		return field.Bool() == (value == "true")

	case reflect.Ptr:
		if field.IsNil() {
			return value == "nil"
		}
		return requireCheckFieldValueFromTop(fl, param, value, defaultNotFoundValue)
	}

	return field.String() == value
}

func IsPhone(phone string) bool {
	const phoneRegex = `\b[+()]?(?<country>1939|1876|1869|1868|1849|1784|1767|1758|1684|1671|1670|1664|1649|1473|1441|1345|1340|1284|1268|1264|1246|1242|998|996|995|994|993|992|977|976|975|974|973|972|971|970|968|967|966|965|964|963|962|961|960|886|880|856|855|853|852|850|692|691|690|689|688|687|686|685|683|682|681|680|679|678|677|676|675|674|673|672|670|598|597|596|595|594|593|592|591|590|509|508|507|506|505|504|503|502|501|500|423|421|420|389|387|386|385|382|381|380|379|378|377|376|375|374|373|372|371|370|359|358|357|356|355|354|353|352|351|350|299|298|297|291|290|269|268|267|266|265|264|263|262|261|260|258|257|256|255|254|253|252|251|250|249|248|246|245|244|243|242|241|240|239|238|237|236|235|234|233|232|231|230|229|228|227|226|225|224|223|222|221|220|218|216|213|212|211|98|95|94|93|92|91|90|86|84|82|81|66|65|64|63|62|61|60|58|57|56|55|54|53|52|51|49|48|47|46|45|44|43|41|40|39|36|34|33|32|31|30|27|20|7|1)[+()]?(?<number>[\d]{4,})\b`
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

func IsEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
