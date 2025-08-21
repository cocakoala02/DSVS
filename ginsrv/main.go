package main

import (
	"fmt"

	"github.com/ldsec/drynx/ginsrv/router"
)

func main() {
	srv := router.ServerSetup()

	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("fail to setup the server: %v", err.Error())
		panic(err)
	}
}
