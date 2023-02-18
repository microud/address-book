package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net"
	"net/http"
)

type Server struct {
	db *Database
}

func NewServer(db *Database) *Server {
	return &Server{db: db}
}

func (s *Server) Run() {
	router := gin.Default()

	router.GET("/address", func(ctx *gin.Context) {
		var address *Address
		var err error
		if ip, ok := ctx.GetQuery("ip"); ok {
			address, err = s.db.FindAddressByIP(ip)
		}

		if mac, ok := ctx.GetQuery("mac"); ok {
			address, err = s.db.FindAddressByMAC(mac)
		}

		if err != nil {
			ctx.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		if address == nil {
			ctx.JSON(500, gin.H{
				"error": "address not found",
			})
			return
		}

		ctx.JSON(200, address)
		return
	})

	router.GET("/address-book", func(ctx *gin.Context) {
		addresses, err := s.db.ListAddresses()
		if err != nil {
			ctx.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(200, addresses)
	})

	server := &http.Server{Handler: router}
	l, err := net.Listen("tcp4", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	err = server.Serve(l)
}
