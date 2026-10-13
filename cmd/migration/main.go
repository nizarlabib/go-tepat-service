package main

import (
	"fmt"
	"log"

	"os"

	"go-memolab-service/internal/entity"
	// "go-memolab-service/internal/entity/master"

	infrastructure2 "go-memolab-service/internal/infrastructure"

	"github.com/spf13/viper"
)

func main() {

	infrastructure2.NewViper()
	db := infrastructure2.NewPostgres(viper.GetString("POSTGRES_CONNECTION"))
	// db.GetClientGorm().Config.DisableForeignKeyConstraintWhenMigrating = true

	enumTypes := []string{
		// `CREATE OR REPLACE FUNCTION "public"."uuid_generate_v4"()
		// RETURNS "pg_catalog"."uuid" AS '$libdir/uuid-ossp', 'uuid_generate_v4'
		// LANGUAGE c VOLATILE STRICT
		// PARALLEL SAFE
		// COST 1;`,
		// `CREATE TYPE gender AS ENUM ('male', 'female');`,
		// `CREATE TYPE user_account_status AS ENUM ('pending', 'suspended', 'verified', 'rejected', 'deleted');`,
	}

	for _, enumSQL := range enumTypes {
		if err := db.GetClientGorm().Exec(enumSQL).Error; err != nil {
			fmt.Errorf("failed to create enum type: %w", err)
			return
		}
	}

	err := db.GetClientGorm().AutoMigrate(
		entity.File{},
		// entity.User{},
		// entity.Role{},
		// entity.DeviceLog{},
		// entity.Permission{},
	)

	if err != nil {
		log.Fatalf("error: %v", err)
		return
	}

	// List SQL file paths
	sqlPaths := []string{
		// "cmd/migration/query.sql",
	}

	// execute SQL statements
	for _, sqlPath := range sqlPaths {
		sqlBytes, err := os.ReadFile(sqlPath)
		if err != nil {
			log.Fatalf("failed to read SQL file: %v", err)
		}

		sqlString := string(sqlBytes)

		if err := db.GetClientGorm().Exec(sqlString).Error; err != nil {
			log.Fatalf("failed to execute SQL: %v", err)
		}
	}

	log.Printf("success")
}
