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
		if err:=views.Init(client); err != nil{
			log.Print("Error:\n")
			log.Fatalln(err.Error())
			return
		}
		router.POST("/record", views.Post)
		for _, field := range views.GetFilteringFields() {
			router.GET("/record/" + field + "/:filter", views.CreateHandler(field, "no"))
			for aggregateFn, aggregateConfig := range views.GetAggregateFnNames() {
				route := "/record/" + field + "/:filter/" + aggregateFn
					if aggregateConfig.HasParam {
					route += "/:param"
				}
				router.GET(
					route,
					views.CreateHandler(field, aggregateFn))
			}
		}

		router.GET("/record/valuesOf/:field", views.ValuesOf)

		router.Run(":3300")
	}
}
