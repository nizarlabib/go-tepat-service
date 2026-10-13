package dd

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/k0kubun/pp/v3"
)

func PrettyPrint(vs ...any) {
	slog.Debug("🚨🚨🚨🚨🚨 PrettyPrint 🚨🚨🚨🚨🚨🚨🚨")
	pp.Println(vs...)
	slog.Debug("🚨🚨🚨🚨🚨 PrettyPrint 🚨🚨🚨🚨🚨🚨🚨")
}

func DiePrettyPrint(vs ...any) {
	slog.Debug("🚨🚨🚨🚨🚨 DiePrettyPrint 🚨🚨🚨🚨🚨🚨🚨")
	pp.Println(vs...)
	slog.Debug("🚨🚨🚨🚨🚨 DiePrettyPrint 🚨🚨🚨🚨🚨🚨🚨")
	panic("stopped cause of DD")
}

func DiePrint(vs ...any) {
	fmt.Println(vs...)
	panic("stopped cause of DD")
}

func DiePrintJson(vs ...any) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("⚠️ Program stopped intentionally (DiePrintJson).")
		}
	}()

	PrintJson(vs...)
	panic("die")
}

func PrintJson(vs ...any) {
	fmt.Println("🚨🚨🚨🚨🚨 DiePrintJson 🚨🚨🚨🚨🚨🚨🚨")
	for _, v := range vs {
		pretty := prepareForJSON(v)
		jsonData, err := json.MarshalIndent(pretty, "", "  ")
		if err != nil {
			fmt.Printf("JSON encode error: %v\n", err)
		} else {
			fmt.Println(string(jsonData))
		}
	}
}
func prepareForJSON(v any) any {
	val := reflect.ValueOf(v)

	if !val.IsValid() || (val.Kind() == reflect.Ptr && val.IsNil()) {
		return nil
	}

	for val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	return val.Interface()
}
