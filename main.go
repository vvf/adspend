package main

import (
	"log"
	"github.com/gin-gonic/gin"
	as "github.com/aerospike/aerospike-client-go"
	"adspend/views"
)

const serverHost = "192.168.77.50"
const serverPort = 3000

func main() {
	// remove timestamps from log messages
	log.SetFlags(0)
	log.Print("Start\n")
	router := gin.Default()

	// connect to the host
	if client, err := as.NewClient(serverHost, serverPort); err != nil {
		log.Print("Error:\n")
		log.Fatalln(err.Error())
	} else {

		recordView := views.RecordView{client, nil}
		recordView.Init()
		router.POST("/record", recordView.Post)
		for _, field := range recordView.GetFilteringFields() {
			router.GET("/record/" + field + "/:filter", recordView.CreateHandler(field, "no"))
			for _, aggregateFn := range recordView.GetAggragateFnNames() {
				router.GET(
					"/record/" + field + "/:filter/" + aggregateFn,
					recordView.CreateHandler(field, aggregateFn))
			}
		}

		router.Run(":3300")
	}
}
