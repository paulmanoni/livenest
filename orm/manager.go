package orm

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	Database string
	Username string
	Password string
	SSLMode  string
}

// Manager wraps GORM with additional functionality
type Manager struct {
	DB     *gorm.DB
	Config *DatabaseConfig
}

// NewManager creates a new ORM manager
func NewManager(config *DatabaseConfig) (*Manager, error) {
	dialector, err := getDialector(config)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Manager{
		DB:     db,
		Config: config,
	}, nil
}

// getDialector returns the appropriate GORM dialector based on config
func getDialector(config *DatabaseConfig) (gorm.Dialector, error) {
	switch config.Driver {
	case "sqlite":
		return sqlite.Open(config.Database), nil

	case "postgres", "postgresql":
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			config.Host,
			config.Username,
			config.Password,
			config.Database,
			config.Port,
			config.SSLMode,
		)
		return postgres.Open(dsn), nil

	case "mysql":
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			config.Username,
			config.Password,
			config.Host,
			config.Port,
			config.Database,
		)
		return mysql.Open(dsn), nil

	default:
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}
}

// AutoMigrate runs auto migration for given models
func (m *Manager) AutoMigrate(models ...interface{}) error {
	return m.DB.AutoMigrate(models...)
}

// Close closes the database connection
func (m *Manager) Close() error {
	sqlDB, err := m.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}