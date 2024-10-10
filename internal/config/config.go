package config

import (
	"log"
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// add this to github

type Config struct {
	Server     ServerConfig
	DB         DBConfig
	Admin      AdminConfig
	SMTP       SMTPConfig
	Cloudinary CloudinaryConfig
	JWT        JWTConfig
	Razorpay   RazorpayConfig
}

type ServerConfig struct {
	Port string
}

type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type AdminConfig struct {
	Username string
	Password string
}

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
}

type JWTConfig struct {
	Secret string
}

type RazorpayConfig struct {
	KeyID     string
	KeySecret string
}

// below code is mainly used in aws, where we load vars from environment
func getEnv(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func Load() (*Config, error) {
	viper.SetConfigName("config") //specify the name of the configuration file (without the extension)
	viper.SetConfigType("yaml")   // specify the format of the configuration file
	viper.AddConfigPath(".")      //add a path where Viper will search for the configuration file
	viper.AddConfigPath("./internal/config")

	err := viper.ReadInConfig() //Reads configuration files
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Println("No config file found. Using environment variables.")
		} else {
			// Config file was found but another error was produced
			log.Printf("Error reading config file: %v", err)
			return nil, err
		}
	} else {
		log.Printf("Config file used: %s", viper.ConfigFileUsed())
	}

	var config Config
	err = viper.Unmarshal(&config) //unmarshal the configuration values from a Viper instance into a struct
	if err != nil {
		log.Printf("Error unmarshaling config: %v", err)
		return nil, err
	}

	// This code is mainly used for aws situation
	// Override with environment variables
	config.Server.Port = getEnv("SERVER_PORT", config.Server.Port)
	config.DB.Host = getEnv("DB_HOST", config.DB.Host)
	config.DB.Port = getEnv("DB_PORT", config.DB.Port)
	config.DB.User = getEnv("DB_USER", config.DB.User)
	config.DB.Password = getEnv("DB_PASSWORD", config.DB.Password)
	config.DB.Name = getEnv("DB_NAME", config.DB.Name)

	config.Admin.Username = getEnv("ADMIN_USERNAME", config.Admin.Username)
	config.Admin.Password = getEnv("ADMIN_PASSWORD", config.Admin.Password)

	config.SMTP.Host = getEnv("SMTP_HOST", config.SMTP.Host)
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", strconv.Itoa(config.SMTP.Port)))
	config.SMTP.Port = smtpPort
	config.SMTP.Username = getEnv("SMTP_USERNAME", config.SMTP.Username)
	config.SMTP.Password = getEnv("SMTP_PASSWORD", config.SMTP.Password)

	config.Cloudinary.CloudName = getEnv("CLOUDINARY_CLOUD_NAME", config.Cloudinary.CloudName)
	config.Cloudinary.APIKey = getEnv("CLOUDINARY_API_KEY", config.Cloudinary.APIKey)
	config.Cloudinary.APISecret = getEnv("CLOUDINARY_API_SECRET", config.Cloudinary.APISecret)

	config.JWT.Secret = getEnv("JWT_SECRET", config.JWT.Secret)

	config.Razorpay.KeyID = getEnv("RAZORPAY_KEY_ID", config.Razorpay.KeyID)
	config.Razorpay.KeySecret = getEnv("RAZORPAY_KEY_SECRET", config.Razorpay.KeySecret)

	return &config, nil
}
