package store

import (
	"os"
	"syscall"

	"github.com/resueman/merch-store/internal/app"
)

const configPath = "./config/config.yaml"

func main() {
	app := app.NewApp(configPath, os.Interrupt, syscall.SIGTERM)
	app.Run()
}
