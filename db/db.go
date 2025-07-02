package db

import (
	"database/sql"
	"fmt"
	"jwt-auth/config"
	"log"
	"net"
	"time"
)

func WaitForTCP(host string, port string, timeout time.Duration) error {
    address := net.JoinHostPort(host, port)
    deadline := time.Now().Add(timeout)

    for {
        conn, err := net.DialTimeout("tcp", address, 2*time.Second)
        if err == nil {
            conn.Close()
            fmt.Printf("Connected to %s\n", address)
            return nil
        }

        if time.Now().After(deadline) {
            return fmt.Errorf("timeout reached, could not connect to %s", address)
        }

        fmt.Printf("Waiting for %s to be available...\n", address)
        time.Sleep(2 * time.Second)
    }
}


func Init(dbConfig *config.DBConfig, DB **sql.DB) {
	var err error
	err = WaitForTCP(dbConfig.Host, dbConfig.Port, 30*time.Second)
	if err != nil {
		log.Fatal("Failed to ping the database port")
	}
	migrationPath := "./migrations"
	config.RunMigrationsUp(migrationPath, dbConfig.DB_URL_string())

	*DB, err = sql.Open("postgres", dbConfig.ConnectionString())
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
	}

	err = (*DB).Ping()
	if err != nil {
		log.Fatal("Database ping failed")
	}

	log.Println("Database connection established")
}

