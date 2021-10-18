package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Client is a custom db client
type Client struct {
	Client *gorm.DB
}

// Ping allows the db to be pinged.
func (c *Client) Ping() error {
	if db, err := c.Client.DB(); err != nil {
		return err
	} else {
		return db.Ping()
	}
}

func (c *Client) Connect() error {
	var err error
	if c.Client, err = gorm.Open(sqlite.Open("../../test/test.db"), &gorm.Config{}); err != nil {
		return err
	} else {
		log.Printf("Connected to %s", c.Client.Config.Name())
		return nil
	}
}
