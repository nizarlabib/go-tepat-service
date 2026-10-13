package convert

import (
	"encoding/base64"
	"errors"
	"fmt"
	"go-memolab-service/internal/value_object/v1/constant"
	"go-memolab-service/internal/value_object/v1/handling"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func Param(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func Time(t *time.Time) *string {
	if t != nil {
		formatted := t.Format(time.DateTime)
		return &formatted
	}

	return nil
}

func GenerateOTP() int64 {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Generate a 4-digit OTP
	otp := rand.Int63n(9000) + 1000 // Generates a number between 1000 and 9999
	return otp
}

func IDUint(r *http.Request, keyNames ...string) uint {
	keyName := "id"
	if len(keyNames) > 0 {
		keyName = keyNames[0]
	}

	ID, err := strconv.ParseUint(r.PathValue(keyName), 10, 32)
	if err != nil {
		return 0
	}
	return uint(ID)
}

func PathValueIDUintChi(r *http.Request) uint {
	ID, err := strconv.ParseUint(r.PathValue("id"), 10, 32)
	if err != nil {
		return 0
	}
	return uint(ID)
}

func PathValueUUIDChi(r *http.Request, name string) (uuid.UUID, error) {
	uuid, err := uuid.Parse(r.PathValue(name))
	if err != nil {
		return uuid, handling.NewHttpError400(err, fmt.Sprintf("%s is not valid", name), constant.ERR_BAD_REQUEST)
	}
	return uuid, nil
}

const base62 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func GenerateString(length int64) string {
	result := make([]byte, length)
	rand.Seed(time.Now().UnixNano())
	for i := range result {
		result[i] = base62[rand.Intn(len(base62))]
	}
	return string(result)
}

func GenerateFileName() string {
	now := time.Now()
	dateNow := now.Format("20060102150405")
	randomString := GenerateString(7)
	result := fmt.Sprintf("%s-%s", dateNow, randomString)
	return result
}

func GenerateFileNameWithSuffix(suffix string) string {
	filename := GenerateFileName()
	result := fmt.Sprintf("%s-%s", filename, suffix)
	return result
}

func GenerateFileNameWithExt(ext string) string {
	filename := GenerateFileName()
	result := fmt.Sprintf("%s.%s", filename, ext)
	return result
}

func TruncateFloat64(f float64, unit float64) float64 {
	bf := big.NewFloat(0).SetPrec(100).SetFloat64(f)
	bu := big.NewFloat(0).SetPrec(100).SetFloat64(unit)

	bf.Quo(bf, bu)

	// Truncate:
	i := big.NewInt(0)
	bf.Int(i)
	bf.SetInt(i)

	f, _ = bf.Mul(bf, bu).Float64()
	return f
}

func UintToStr(v uint) string {
	return fmt.Sprintf("%d", v)
}

func StrToUint(v string) (uint, error) {
	parsed, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func Float64ToStr(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func PointerStr(v string) *string {
	return &v
}

func StringToUUID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

func UUIDToString(id uuid.UUID) string {
	return id.String()
}

func EncodeFileToBase64(path string) (string, error) {
	// Baca semua isi file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Encode ke base64
	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded, nil
}

func ConvertUTMToLatLon(easting, northing float64, zoneNumber int, zoneLetter string) (float64, float64, error) {
	if zoneNumber < 1 || zoneNumber > 60 {
		return 0, 0, errors.New("zone number out of range")
	}

	a := 6378137.0
	eccSquared := 0.00669438
	k0 := 0.9996

	x := easting - 500000.0
	y := northing
	if rune(zoneLetter[0]) == 'S' {
		y -= 10000000.0
	}

	longOrigin := float64((zoneNumber-1)*6 - 180 + 3)
	eccPrimeSquared := eccSquared / (1 - eccSquared)

	M := y / k0
	mu := M / (a * (1 - eccSquared/4.0 - 3*eccSquared*eccSquared/64.0 - 5*eccSquared*eccSquared*eccSquared/256.0))

	e1 := (1 - math.Sqrt(1-eccSquared)) / (1 + math.Sqrt(1-eccSquared))
	J1 := 3*e1/2 - 27*math.Pow(e1, 3)/32.0
	J2 := 21*e1*e1/16 - 55*math.Pow(e1, 4)/32.0
	J3 := 151 * math.Pow(e1, 3) / 96.0
	J4 := 1097 * math.Pow(e1, 4) / 512.0

	fp := mu + J1*math.Sin(2*mu) + J2*math.Sin(4*mu) + J3*math.Sin(6*mu) + J4*math.Sin(8*mu)

	C1 := eccPrimeSquared * math.Pow(math.Cos(fp), 2)
	T1 := math.Pow(math.Tan(fp), 2)
	R1 := a * (1 - eccSquared) / math.Pow(1-eccSquared*math.Pow(math.Sin(fp), 2), 1.5)
	N1 := a / math.Sqrt(1-eccSquared*math.Pow(math.Sin(fp), 2))

	D := x / (N1 * k0)

	lat := fp - (N1*math.Tan(fp)/R1)*(D*D/2-
		(5+3*T1+10*C1-4*C1*C1-9*eccPrimeSquared)*math.Pow(D, 4)/24+
		(61+90*T1+298*C1+45*T1*T1-252*eccPrimeSquared-3*C1*C1)*math.Pow(D, 6)/720)
	lat = lat * (180 / math.Pi)

	lon := (D - (1+2*T1+C1)*math.Pow(D, 3)/6 +
		(5-2*C1+28*T1-3*C1*C1+8*eccPrimeSquared+24*T1*T1)*math.Pow(D, 5)/120) / math.Cos(fp)
	lon = longOrigin + lon*(180/math.Pi)

	return lat, lon, nil
}

func ToPointerStr(v string) *string {
	return ToPointer(v)
}

func ToPointerBool(b bool) *bool {
	return ToPointer(b)
}

func ToPointerTime(v time.Time) *time.Time {
	return ToPointer(v)
}

func ToPointer[T any](v T) *T {
	switch any(v).(type) {
	case string, int, int64, float64, bool, time.Time:
		// These types can be directly converted to a pointer
		return &v
	default:
		// For other types, we need to create a new instance
		newValue := new(T)
		*newValue = v
		return newValue
	}
}

func ValidateUUID(id, name string, required bool) (uuid.UUID, error) {
	if id == "" {
		if required {
			return uuid.UUID{}, handling.NewHttpError400(nil, fmt.Sprintf("%s is required", name), constant.ERR_BAD_REQUEST)
		}
		return uuid.UUID{}, nil
	}
	parsedUUID, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, handling.NewHttpError400(err, fmt.Sprintf("%s is not valid", name), constant.ERR_BAD_REQUEST)
	}
	return parsedUUID, nil
}

func Round2(val float64) float64 {
	return math.Round((val+1e-9)*100) / 100
}

func RoundFloat32(val float32) float32 {
	// step 1: *1000 untuk ambil digit ketiga
	raw := float64(val) * 1000

	// step 2: bulatkan ke integer buat menghilangkan floating-point noise
	valInt := int64(math.Round(raw))

	// step 3: ambil digit ke-3
	thirdDigit := valInt % 10

	// step 4: hilangkan digit ke-3
	valInt /= 10

	// step 5: rounding standar 5-9 naik, 0-4 turun
	if thirdDigit >= 5 {
		valInt++
	}

	// step 6: konversi kembali ke dua desimal
	return float32(float64(valInt) / 100)
}

func Base64Encode(v []byte) string {
	return base64.StdEncoding.EncodeToString(v)
}

func Base64Decode(v string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, err
	}
	return decoded, nil
}

func BoolToStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
