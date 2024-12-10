package control

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

func CaptureSigTerm() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Print("\n")
		log.Info().Msg("(╯°□°)╯︵ ┻━━┻")

		os.Exit(1)
	}()
}
