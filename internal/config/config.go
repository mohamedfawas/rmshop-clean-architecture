package config

import (
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig
	DB         DBConfig
	Admin      AdminConfig
	SMTP       SMTPConfig
	Cloudinary CloudinaryConfig
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

func Load() (*Config, error) {
	viper.SetConfigName("config") //specify the name of the configuration file (without the extension)
	viper.SetConfigType("yaml")   // specify the format of the configuration file
	viper.AddConfigPath(".")      //add a path where Viper will search for the configuration file
	viper.AddConfigPath("./internal/config")

	err := viper.ReadInConfig() //Reads configuration files
	if err != nil {
		log.Printf("Error reading config file: %v", err)
		return nil, err
	}

	log.Printf("Config file used: %s", viper.ConfigFileUsed())

	var config Config
	err = viper.Unmarshal(&config) //unmarshal the configuration values from a Viper instance into a struct
	if err != nil {
		log.Printf("Error unmarshaling config: %v", err)
		return nil, err
	}
	log.Printf("values of db are : %s", config.DB.Host)
	log.Printf("values of cloudinary are : %s", config.Cloudinary.CloudName)

	// Debug print Cloudinary configuration
	log.Printf("Loaded Cloudinary config: %+v", config.Cloudinary)
	return &config, nil
}
