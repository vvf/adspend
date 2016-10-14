package main

import (
	"log"
	"time"
	"github.com/gin-gonic/gin"
	as "github.com/aerospike/aerospike-client-go"
	"net/http"
	"webformula/models"
)

var serverHost = "192.168.77.50"
var serverPort = 3000
var namespace = "test"
var setName = "adspend"
var binName = "traffic"

const gig int64 = 1000000000

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
		// Initialize policy.
		policy := as.NewWritePolicy(0, 0)
		policy.Timeout = 50 * time.Millisecond  // 50 millisecond timeout.
		policy.Expiration = 3600 * 24 * 60 // 60 days NB! this is time how long records will be store

		router.POST("/record", func(c *gin.Context) {
			var response struct {
				Success bool
				Message string
			}
			var record models.BidRequest
			if c.BindJSON(&record) == nil {
				record.Timestamp = time.Now()
				key, _ := as.NewKey(namespace, setName, record.Timestamp.Format(time.ANSIC))

				if err := client.PutObject(policy, key, &record); err == nil {
					response.Success = true
				} else {
					response.Success = false
					response.Message = "Error in put bin:" + string(err.Error())
				}
				c.JSON(http.StatusOK, response)
			} else {
				response.Success = false
				response.Message = ""
				c.JSON(http.StatusNotAcceptable, response)
			}
		})
		router.GET("/record", func(c *gin.Context) {
			results := make([]interface{}, 0, 20)

			stmt := as.NewStatement(namespace, setName)
			asResults := make(chan *models.BidRequest, 10)

			tsTo := time.Now().Unix()
			tsFrom := tsTo - 600
			stmt.Addfilter(as.NewRangeFilter("Timestamp", int64(tsFrom) * gig, int64(tsTo+1) * gig))
			_, err := client.QueryObjects(nil, stmt, asResults)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			} else {
				for res := range asResults {
					pk := len(results)
					results = results[:pk + 1]
					results[pk] = res
				}
				c.JSON(http.StatusOK, gin.H{"results": results, "range": []int64{tsFrom, tsTo}})
			}
		})

		router.Run(":3300")
	}
}
