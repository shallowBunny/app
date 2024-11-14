package utils

import (
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

func IsLocalhostTesting() bool {
	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to get hostname")
	}
	return hostname == os.Getenv("SHALLOWBUNNY_LOCALHOST")
}

func VerifyFilePath(photoFilePath string) error {
	_, err := os.Stat(photoFilePath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file does not exist at path: %s", photoFilePath)
	}
	if err != nil {
		return fmt.Errorf("file path error: %s, %v", photoFilePath, err)
	}
	return nil
}
