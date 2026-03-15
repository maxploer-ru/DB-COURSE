package config

type Config struct {
	Port     int
	Database struct {
		User     string
		Password string
		Name     string
		Host     string
		Port     int
	}
}
