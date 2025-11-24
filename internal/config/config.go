package config

import (
	"log"
	"strings"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	DB         DBConfig         `mapstructure:"db"`
	Admin      AdminConfig      `mapstructure:"admin"`
	SMTP       SMTPConfig       `mapstructure:"smtp"`
	Cloudinary CloudinaryConfig `mapstructure:"cloudinary"`
	JWT        JWTConfig        `mapstructure:"jwt"`
	Razorpay   RazorpayConfig   `mapstructure:"razorpay"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type AdminConfig struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type SMTPConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	FromEmail string `mapstructure:"from_email"`
	FromName  string `mapstructure:"from_name"`
}

type CloudinaryConfig struct {
	CloudName string `mapstructure:"cloud_name"`
	APIKey    string `mapstructure:"api_key"`
	APISecret string `mapstructure:"api_secret"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type RazorpayConfig struct {
	KeyID     string `mapstructure:"key_id"`
	KeySecret string `mapstructure:"key_secret"`
}

func Load() (*Config, error) {
	// Load .env into process env (non-fatal if file is missing)
	_ = gotenv.Load()

	v := viper.New()

	// ENV vars override everything
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	bindEnvVars(v)
	setDefaults(v)

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Printf("Error unmarshaling config: %v", err)
		return nil, err
	}

	return &cfg, nil
}

func bindEnvVars(v *viper.Viper) {
	keys := []string{
		"server.port",

		"db.host",
		"db.port",
		"db.user",
		"db.password",
		"db.name",

		"admin.username",
		"admin.password",

		"smtp.host",
		"smtp.port",
		"smtp.username",
		"smtp.password",
		"smtp.from_email",
		"smtp.from_name",

		"cloudinary.cloud_name",
		"cloudinary.api_key",
		"cloudinary.api_secret",

		"jwt.secret",

		"razorpay.key_id",
		"razorpay.key_secret",
	}

	for _, key := range keys {
		_ = v.BindEnv(key)
	}
}

func setDefaults(v *viper.Viper) {
	// Server
	v.SetDefault("server.port", "8080")

	// Database
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", "5432")
	v.SetDefault("db.user", "postgres")
	v.SetDefault("db.password", "postgres")
	v.SetDefault("db.name", "rmshop_db")

	// Admin
	v.SetDefault("admin.username", "admin")
	v.SetDefault("admin.password", "admin@123")

	// SMTP
	v.SetDefault("smtp.host", "smtp.gmail.com")
	v.SetDefault("smtp.port", 587)
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from_email", "")
	v.SetDefault("smtp.from_name", "Real Madrid Shop")

	// Cloudinary
	v.SetDefault("cloudinary.cloud_name", "")
	v.SetDefault("cloudinary.api_key", "")
	v.SetDefault("cloudinary.api_secret", "")

	// JWT
	v.SetDefault("jwt.secret", "change-me-in-production")

	// Razorpay
	v.SetDefault("razorpay.key_id", "")
	v.SetDefault("razorpay.key_secret", "")
}
