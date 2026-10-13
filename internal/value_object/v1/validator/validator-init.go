package validator

import (
	"fmt"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"
	"net/http"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Validator function to validate Request struct fields
func validator(e any) error {
	v := reflect.ValueOf(e)
	t := reflect.TypeOf(e)

	for i := range v.NumField() {
		field := v.Field(i)
		fieldType := t.Field(i)
		jsonTag := fieldType.Tag.Get("json")
		tag := fieldType.Tag.Get("validate")
		isOptional := strings.Contains(tag, "optional")

		// Check for required field
		if strings.Contains(tag, "required_if") {
			parts := strings.Split(tag, ",")
			for _, part := range parts {
				if strings.HasPrefix(part, "required_if:") {
					conds := strings.TrimPrefix(part, "required_if:")
					x := strings.Split(conds, "=")
					if len(x) > 1 {
						if v.FieldByName(x[0]).String() == x[1] && !isRequired(field) {
							return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s is required while %s is %s", jsonTag, x[0], x[1]), constant.ERR_VALIDATION_ERROR)
						} else {
							isOptional = true
						}
					}
				}
			}
		} else if strings.Contains(tag, "required") {
			if !isRequired(field) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s is required", jsonTag), constant.ERR_VALIDATION_ERROR)
			}
		}

		// Check for min length
		if strings.Contains(tag, "min=") {
			minValue := extractMinValue(tag)
			if len(field.String()) < minValue {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s minimal harus %d karakter", jsonTag, minValue), constant.ERR_VALIDATION_ERROR)
			}
		}

		// Check for max length
		if strings.Contains(tag, "max=") {
			maxValue := extractMaxValue(tag)
			if len(field.String()) > maxValue {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s maksimal harus %d karakter", jsonTag, maxValue), constant.ERR_VALIDATION_ERROR)
			}
		}

		// Check for valid email
		if strings.Contains(tag, "email") {
			if isOptional && isEmpty(field.String()) {
				continue
			}
			if !isEmail(field.String()) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s format email tidak valid", jsonTag), constant.ERR_VALIDATION_ERROR)
			}
		}

		// Check for age >= specified minimum (default is 1 if gte is present without value)
		if strings.Contains(tag, "gte") {
			minValue := extractGteValue(tag)
			if field.Kind() == reflect.Int64 && field.Int() < int64(minValue) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s minimal harus %d", jsonTag, minValue), constant.ERR_VALIDATION_ERROR)
			}
		}

		if strings.Contains(tag, "lte") {
			minValue := extractLteValue(tag)
			if field.Kind() == reflect.Int64 && field.Int() > int64(minValue) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s maximal %d", jsonTag, minValue), constant.ERR_VALIDATION_ERROR)
			}
		}

		if strings.Contains(tag, "password") {
			password := field.String()
			level := PasswordStrengthMeter(password)
			if Terrible == level {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s kurang kuat", jsonTag), constant.ERR_VALIDATION_ERROR)
			}
		}

		if strings.Contains(tag, "phone") {
			if !isPhone(field.String()) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s format tidak valid", jsonTag), constant.ERR_VALIDATION_ERROR)
			}
		}

		if strings.Contains(tag, "number") {
			if !isNumberString(field.String()) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s bukan angka", jsonTag), constant.ERR_VALIDATION_ERROR)
			}
		}

		if strings.Contains(tag, "date") {
			if !isDate(field.String()) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s format tidak valid DD-MM-YYYY", jsonTag), constant.ERR_VALIDATION_ERROR)
			}
		}

		if strings.Contains(tag, "len=") {
			lenValue := extractLenValue(tag)
			if len(field.String()) != lenValue {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s harus %d karakter", jsonTag, lenValue), constant.ERR_VALIDATION_ERROR)
			}
		}

		if strings.Contains(tag, "enum=") {
			if isOptional && isEmpty(field.String()) {
				continue
			}
			enumValues := extractEnumValue(tag)
			if !slices.Contains(enumValues, field.String()) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s harus salah satu dari: %s", jsonTag, strings.Join(enumValues[:], ",")), constant.ERR_VALIDATION_ERROR)
			}
		}

		if strings.Contains(tag, "uuid") {
			if isOptional && isEmpty(field.String()) {
				continue
			}
			if !isUUID(field.String()) {
				return handling.NewHttpError(nil, http.StatusBadRequest, fmt.Sprintf("%s bukan uuid", jsonTag), constant.ERR_VALIDATION_ERROR)
			}
		}
	}

	return nil // Return nil if all validations pass
}

// Helper function to check if a string is empty or contains only whitespace
func isEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

func isDate(date string) bool {
	_, err := time.Parse("02-01-2006", date)
	return err == nil
}

func isNumberString(numberString string) bool {
	_, err := strconv.Atoi(numberString)
	return err == nil
}

func isPhone(phone string) bool {
	const phoneRegex = `\b[+()]?(?<country>1939|1876|1869|1868|1849|1784|1767|1758|1684|1671|1670|1664|1649|1473|1441|1345|1340|1284|1268|1264|1246|1242|998|996|995|994|993|992|977|976|975|974|973|972|971|970|968|967|966|965|964|963|962|961|960|886|880|856|855|853|852|850|692|691|690|689|688|687|686|685|683|682|681|680|679|678|677|676|675|674|673|672|670|598|597|596|595|594|593|592|591|590|509|508|507|506|505|504|503|502|501|500|423|421|420|389|387|386|385|382|381|380|379|378|377|376|375|374|373|372|371|370|359|358|357|356|355|354|353|352|351|350|299|298|297|291|290|269|268|267|266|265|264|263|262|261|260|258|257|256|255|254|253|252|251|250|249|248|246|245|244|243|242|241|240|239|238|237|236|235|234|233|232|231|230|229|228|227|226|225|224|223|222|221|220|218|216|213|212|211|98|95|94|93|92|91|90|86|84|82|81|66|65|64|63|62|61|60|58|57|56|55|54|53|52|51|49|48|47|46|45|44|43|41|40|39|36|34|33|32|31|30|27|20|7|1)[+()]?(?<number>[\d]{4,})\b`
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

// Helper function to validate email format
func isEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}

func isUUID(v string) bool {
	_, err := uuid.Parse(v)
	return err == nil
}

func isRequired(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.String:
		if isEmpty(field.String()) {
			return false
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Int() == 0 { // 0 dianggap kosong untuk integer
			return false
		}
	case reflect.Float32, reflect.Float64:
		if field.Float() == 0.0 { // 0.0 dianggap kosong untuk float
			return false
		}
	case reflect.Slice, reflect.Array:
		if field.Len() == 0 { // Panjang array/slice kosong
			return false
		}
	case reflect.Ptr:
		if field.IsNil() { // Pointer nil
			return false
		}
	default:
		if !field.IsValid() || field.IsZero() { // Nilai default dianggap kosong
			return false
		}
	}
	return true
}

// Helper function to extract minimum value from tag
func extractMinValue(tag string) int {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "min=") {
			minStr := strings.TrimPrefix(part, "min=")
			minValue, err := strconv.Atoi(minStr)
			if err == nil {
				return minValue
			}
		}
	}
	return 0 // Return 0 if not found
}

// Helper function to extract maximum value from tag
func extractMaxValue(tag string) int {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "max=") {
			maxStr := strings.TrimPrefix(part, "max=")
			maxValue, err := strconv.Atoi(maxStr)
			if err == nil {
				return maxValue
			}
		}
	}
	return 0 // Return 0 if not found
}

// Helper function to extract gte value from tag | "greater than or equal to" atau "lebih besar atau sama dengan."
func extractGteValue(tag string) int {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "gte=") {
			gteStr := strings.TrimPrefix(part, "gte=")
			gteValue, err := strconv.Atoi(gteStr)
			if err == nil {
				return gteValue
			}
		}
	}
	return 1 // Default to 1 if gte is present without a specified value
}

func extractLenValue(tag string) int {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "len=") {
			gteStr := strings.TrimPrefix(part, "len=")
			gteValue, err := strconv.Atoi(gteStr)
			if err == nil {
				return gteValue
			}
		}
	}
	return 1 // Default to 1 if gte is present without a specified value
}

// Helper function to extract lte value from tag | "less than or equal to" (kurang dari atau sama dengan
func extractLteValue(tag string) int {
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "lte=") {
			lteStr := strings.TrimPrefix(part, "lte=")
			lteValue, err := strconv.Atoi(lteStr)
			if err == nil {
				return lteValue
			}
		}
	}
	return 0 // Default to 0 if lte is not found
}

func extractEnumValue(tag string) []string {
	values := []string{}
	parts := strings.Split(tag, ",")
	for _, part := range parts {
		if strings.HasPrefix(part, "enum=") {
			enums := strings.Split(strings.TrimPrefix(part, "enum="), "|")
			values = append(values, enums...)
		}
	}
	return values
}

func IsPhone(phone string) bool {
	const phoneRegex = `\b[+()]?(?<country>1939|1876|1869|1868|1849|1784|1767|1758|1684|1671|1670|1664|1649|1473|1441|1345|1340|1284|1268|1264|1246|1242|998|996|995|994|993|992|977|976|975|974|973|972|971|970|968|967|966|965|964|963|962|961|960|886|880|856|855|853|852|850|692|691|690|689|688|687|686|685|683|682|681|680|679|678|677|676|675|674|673|672|670|598|597|596|595|594|593|592|591|590|509|508|507|506|505|504|503|502|501|500|423|421|420|389|387|386|385|382|381|380|379|378|377|376|375|374|373|372|371|370|359|358|357|356|355|354|353|352|351|350|299|298|297|291|290|269|268|267|266|265|264|263|262|261|260|258|257|256|255|254|253|252|251|250|249|248|246|245|244|243|242|241|240|239|238|237|236|235|234|233|232|231|230|229|228|227|226|225|224|223|222|221|220|218|216|213|212|211|98|95|94|93|92|91|90|86|84|82|81|66|65|64|63|62|61|60|58|57|56|55|54|53|52|51|49|48|47|46|45|44|43|41|40|39|36|34|33|32|31|30|27|20|7|1)[+()]?(?<number>[\d]{4,})\b`
	re := regexp.MustCompile(phoneRegex)
	return re.MatchString(phone)
}

// Helper function to validate email format
func IsEmail(email string) bool {
	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	return re.MatchString(email)
}
