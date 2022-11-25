package Services

import (
	"github.com/joho/godotenv"
	"log"
)

func DotEnvInitialize() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Couldn't initialize DotEnv ( " + err.Error() + " ). Fallback to native.")
	}
}
