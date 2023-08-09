package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"

	"database/sql"
	"github.com/hibiken/asynq"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var can_schedule bool
var currently_schedules bool

func main() {

	envLoadErr := godotenv.Load()
	if envLoadErr != nil {
		log.Println(envLoadErr)
	}

	can_schedule_env, csPresent := os.LookupEnv("CAN_SCHEDULE")
	if !csPresent {
		can_schedule = false
	} else {
		cs, err := strconv.ParseBool(can_schedule_env)
		if err != nil {
			log.Fatal(err)
		}
		can_schedule = cs
	}

	db := setupDB()
	defer db.Close()
	if can_schedule {
		go start_scheduler()
	}

	redisConnOpt := asynq.RedisClientOpt{
		Addr: "localhost:6379",
		// Omit if no password is required
		// Password: "mypassword",
		// Use a dedicated db number for asynq.
		// By default, Redis offers 16 databases (0..15)
		DB: 0,
	}

	srv := asynq.NewServer(
		redisConnOpt,
		asynq.Config{Concurrency: 10},
	)

	if err := srv.Run(asynq.HandlerFunc(task_handler)); err != nil {
		log.Fatal(err)
	}
}

func task_handler(ctx context.Context, task *asynq.Task) error {
	switch task.Type() {
	case "wasm":
		fmt.Println("handeling wasm message")
		var wasmTask WasmTask
		err := json.Unmarshal(task.Payload(), &wasmTask)
		if err != nil {
			return err
		}

		fmt.Println(wasmTask)
	case "docker":
		fmt.Println("handleling docker message")
		var dockerTask DockerTask
		err := json.Unmarshal(task.Payload(), &dockerTask)
		if err != nil {
			return err
		}
	}
	return nil
}

func setupDB() *sql.DB {
	dbUser, uPresent := os.LookupEnv("DB_USER")
	if !uPresent {
		log.Fatal("DB_USER enviromental variable not set")
	}

	dbPasswd, pPresent := os.LookupEnv("DB_PASSWD")
	if !pPresent {
		log.Fatal("DB_PASSWD enviromental variable not set")
	}

	dbUrl, dbPresent := os.LookupEnv("DB_URL")
	if !dbPresent {
		log.Fatal("DB_URL enviromental variable not set")
	}

	dbName, namePresent := os.LookupEnv("DB_NAME")
	if !namePresent {
		log.Fatal("DB_NAME enviromental variable not set")
	}

	fmt.Printf("connecting to DB %v at %v as %v:%v\n", dbName, dbUrl, dbUser, dbPasswd)

	connStr := fmt.Sprintf("user=%v password=%v dbname=%v host=%v sslmode=verify-full", dbUser, dbPasswd, dbName, dbUrl)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error connectiong to DB.")
		panic(err)
	}

	var version string
	if err := db.QueryRow("select version();").Scan(&version); err != nil {
		panic(err)
	}

	fmt.Printf("version=%s\n", version)
	return db
}

func start_scheduler(db sql.DB) {

}
