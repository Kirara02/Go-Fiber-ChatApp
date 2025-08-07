package scheduler

import (
	"log"
	"main/repository"

	"github.com/robfig/cron/v3"
)

func InitCronJobs(tokenRepo repository.TokenRepository) *cron.Cron {
	c := cron.New()

	_, err := c.AddFunc("@daily", func() {
		log.Println("Running scheduled job: Cleaning expired tokens...")

		rowsAffected, err := tokenRepo.CleanExpiredTokens()
		if err != nil {
			log.Printf("Error cleaning expired tokens: %v", err)
			return
		}

		log.Printf("Successfully cleaned expired tokens. Rows affected: %d", rowsAffected)
	})
	if err != nil {
		log.Fatalf("Could not add cron job: %v", err)
	}

	log.Println("Cron job scheduler started.")
	c.Start()

	return c
}
