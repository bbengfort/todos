package todos

import (
	"time"

	"github.com/jinzhu/gorm"
)

const (
	tokenCleanupSvcInterval = 1 * time.Hour
)

// TokensCleanupService is a go routine that runs the TokenCleanup function every hour,
// logging the results to disk. It is run when it is first called, then every hour.
func (s *API) TokensCleanupService() {
	logger.Printf("starting token cleanup service")
	ticker := time.NewTicker(tokenCleanupSvcInterval)

	for {
		// Execute the tokens cleanup command and log the results
		rows, err := TokensCleanup(s.db)
		if err != nil {
			logger.Printf("cleaned up %d rows from tokens before error: %s", rows, err)
		} else if rows > 0 {
			logger.Printf("cleaned up %d expired tokens from the database", rows)
		} else {
			// TODO: make this a lower log level when log leveling is a thing.
			logger.Printf("no expired tokens to clean up")
		}

		// Block until the next scheduled service run
		<-ticker.C
	}
}

// TokensCleanup iterates through the database and finds any tokens that have expired,
// deleting them from the database. It returns the number of rows deleted and any
// errors that might have occurred during processing. Note that this function is run
// inside of a transaction in case it is long running.
func TokensCleanup(db *gorm.DB) (rows int, err error) {
	err = db.Transaction(func(tx *gorm.DB) (err error) {
		now := time.Now()
		var tokens []Token
		if err = tx.Where("refresh_by < ?", now).Find(&tokens).Error; err != nil {
			return err
		}

		for _, token := range tokens {
			// Secondary check is probably unnecessary
			if token.RefreshBy.Before(now) {
				if err = tx.Delete(&token).Error; err != nil {
					return err
				}
				rows++
			}
		}

		return nil // will commit the transaction
	})
	return rows, err
}
