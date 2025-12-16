package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

func Logger() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func AuditLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()

		userID := "-"
		role := "-"

		if uid := c.Locals("user_id"); uid != nil {
			userID = fmt.Sprintf("%v", uid)
		}
		if r := c.Locals("role"); r != nil {
			role = fmt.Sprintf("%v", r)
		}

		log.Printf("| user_id=%s | role=%s | %s %s | %d | %v",
			userID,
			role,
			c.Method(),
			c.Path(),
			c.Response().StatusCode(),
			time.Since(start),
		)

		return err
	}
}
