package config

import "github.com/spf13/viper"

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Admin  AdminConfig
	SMTP   SMTPConfig
}

type ServerConfig struct {
	Port string
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

	// Set defaults for SMTP config
	viper.SetDefault("SMTP.Host", "smtp.example.com")
	viper.SetDefault("SMTP.Port", 587)
	viper.SetDefault("SMTP.Username", "your_username")
	viper.SetDefault("SMTP.Password", "your_password")

	err := viper.ReadInConfig() //Reads configuration files
	if err != nil {
		return nil, err
	}

	var config Config
	err = viper.Unmarshal(&config) //unmarshal the configuration values from a Viper instance into a struct
	if err != nil {
		return nil, err
	}

	return &config, nil
}
