package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDatabase_UpsertAddress(t *testing.T) {
	db, err := NewDatabase()
	assert.Nil(t, err)

	err = db.SaveAddress(Address{
		IP:  "192.168.16.1",
		MAC: "18:3e:ef:d0:cc:58",
	})
	assert.Nil(t, err)

	err = db.SaveAddress(Address{
		IP:  "192.168.16.12",
		MAC: "18:3e:ef:d0:cc:591",
	})
	assert.Nil(t, err)
}
