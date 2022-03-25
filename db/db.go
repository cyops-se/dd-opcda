package db

import (
	"dd-opcda/types"
	"log"
	"path"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase(ctx types.Context) {
	filename := path.Join(ctx.Wdir, "test.db")
	database, err := gorm.Open(sqlite.Open(filename), &gorm.Config{})
	// dsn := "user=dev password=hemligt dbname=dev host=localhost port=5432"
	// database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %s", err.Error())
		return
	}

	log.Printf("Database connected")

	// The User model is special due to the 'password' field, and has
	// user specific routes
	database.AutoMigrate(&types.User{})

	// Generic CRUD data types
	ConfigureTypes(database, types.Log{}, types.KeyValuePair{})
	ConfigureTypes(database, types.User{}, types.Settings{})
	ConfigureTypes(database, types.DiodeProxy{})
	ConfigureTypes(database, types.OPCGroup{}, types.OPCTag{})
	ConfigureTypes(database, types.FileTransferConfig{})

	DB = database
}

func ConfigureTypes(database *gorm.DB, datatypes ...interface{}) {
	for _, datatype := range datatypes {
		stmt := &gorm.Statement{DB: database}
		stmt.Parse(datatype)
		name := stmt.Schema.Table
		types.RegisterType(name, datatype)
		database.AutoMigrate(datatype)
	}
}
