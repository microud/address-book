package main

import (
	"github.com/hashicorp/go-memdb"
	"github.com/rs/xid"
)

type Address struct {
	ID  string `json:"id"`
	IP  string `json:"ip"`
	MAC string `json:"mac"`
}

type Database struct {
	*memdb.MemDB
}

func NewDatabase() (*Database, error) {
	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"address": {
				Name: "address",
				Indexes: map[string]*memdb.IndexSchema{
					"id": {
						Name:         "id",
						Unique:       true,
						AllowMissing: false,
						Indexer:      &memdb.StringFieldIndex{Field: "ID"},
					},
					"ip": {
						Name:         "ip",
						Unique:       true,
						AllowMissing: false,
						Indexer:      &memdb.StringFieldIndex{Field: "IP"},
					},
					"mac": {
						Name:         "mac",
						Unique:       true,
						AllowMissing: false,
						Indexer:      &memdb.StringFieldIndex{Field: "MAC"},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		return nil, err
	}

	return &Database{MemDB: db}, nil
}

func (d *Database) FindAddressByIP(ip string) (*Address, error) {
	txn := d.MemDB.Txn(false)
	defer txn.Abort()
	raw, err := txn.First("address", "ip", ip)
	if err != nil {
		return nil, err
	}

	if raw == nil {
		return nil, nil
	}

	return raw.(*Address), nil
}

func (d *Database) FindAddressByMAC(mac string) (*Address, error) {
	txn := d.MemDB.Txn(false)
	defer txn.Abort()
	raw, err := txn.First("address", "mac", mac)
	if err != nil {
		return nil, err
	}

	if raw == nil {
		return nil, nil
	}

	return raw.(*Address), nil
}

func (d *Database) SaveAddress(address Address) error {
	txn := d.MemDB.Txn(true)
	defer txn.Abort()
	raw, err := txn.First("address", "ip", address.IP)
	if err != nil {
		return err
	}

	if raw == nil {
		raw, err = txn.First("address", "mac", address.MAC)
		if err != nil {
			return err
		}
	}

	if raw != nil {
		address.ID = raw.(*Address).ID
	} else {
		address.ID = xid.New().String()
	}

	err = txn.Insert("address", &address)
	if err != nil {
		return err
	}

	txn.Commit()
	return nil
}

func (d *Database) ListAddresses() ([]*Address, error) {
	var addresses []*Address
	txn := d.Txn(false)

	it, err := txn.Get("address", "id")
	if err != nil {
		return nil, err
	}

	for raw := it.Next(); raw != nil; raw = it.Next() {
		addresses = append(addresses, raw.(*Address))
	}

	return addresses, nil
}
