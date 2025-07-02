package config

func GetServerPort() string {
	return GetEnv("SERVER_PORT")
}
