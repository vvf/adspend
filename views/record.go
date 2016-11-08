package views

import (
	"github.com/gin-gonic/gin"
	"adspend/models"
	"time"
	"net/http"
	as "github.com/aerospike/aerospike-client-go"
	"strings"
	"strconv"
	"os"
	"github.com/aerospike/aerospike-client-go/logger"
)

const namespace = "test"
const setName = "adspend"

const LIMIT = 1000
const GIG int64 = 1000000000

var client *as.Client
var policy *as.WritePolicy

func Init(c *as.Client) error {
	client = c
	policy = as.NewWritePolicy(0, 0)
	policy.Timeout = 50 * time.Millisecond  // 50 millisecond timeout.
	policy.Expiration = 3600 * 24 * 60 // 60 days NB! this is time how long records will be store

	// create indexes
	for _, field := range GetFilteringFields() {
		task, err := client.CreateIndex(policy, namespace, setName, field + "Index", field, as.STRING)
		if err == nil {
			err = <-task.OnComplete()
		}
		if err != nil {
			logger.Logger.Logger.Printf(err.Error())
		}
	}

	// register UDF
	luaPath, _ := os.Getwd()
	luaPath += "/src/adspend/udfs/"
	as.SetLuaPath(luaPath)
	task, err := client.RegisterUDFFromFile(policy, luaPath + "record_udfs.lua", "record_udfs.lua", as.LUA)
	if err != nil {
		return err
	}
	<-task.OnComplete()
	return nil
}

func Post(c *gin.Context) {
	var response struct {
		Success bool
		Message string
	}
	var record models.BidRequest
	if err := c.BindJSON(&record); err == nil {
		record.Timestamp = time.Now()
		key, _ := as.NewKey(namespace, setName, record.Timestamp.Format(time.ANSIC))

		if err := client.PutObject(policy, key, &record); err == nil {
			response.Success = true
			c.JSON(http.StatusOK, response)
		} else {
			response.Success = false
			response.Message = "Error in put bin:" + string(err.Error())
			c.JSON(http.StatusInternalServerError, response)
		}
	} else {
		response.Success = false
		response.Message = err.Error()
		c.JSON(http.StatusNotAcceptable, response)
	}
}

func addRangeFilter(stmt *as.Statement, field string, filterBy string) error {
	filterValues := strings.Split(filterBy, "-")
	var tsFrom int64
	var tsTo int64
	if filterValues[0] != "" {
		if ts, err := strconv.Atoi(filterValues[0]); err != nil {
			return err
		} else {
			tsFrom = int64(ts)
		}
	} else {
		tsFrom = 0
	}
	if filterValues[1] != "" {
		if ts, err := strconv.Atoi(filterValues[1]); err != nil {
			return err
		} else {
			tsTo = int64(ts)
		}
	} else {
		tsTo = 0xFFFFFFFF
	}

	stmt.Addfilter(as.NewRangeFilter(field, tsFrom * GIG, (tsTo + 1) * GIG))
	return nil
}

func CreateHandler(field string, aggregateFnName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := c.Param("filter")
		result := gin.H{"field": field, "filterValue": filter, "ts": time.Now().Unix() }
		stmt := as.NewStatement(namespace, setName)
		asResults := make(chan *models.BidRequest, 10)
		if strings.Contains(filter, "-") && field == "Timestamp" {
			if err := addRangeFilter(stmt, field, filter); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
		} else {
			stmt.Addfilter(as.NewEqualFilter(field, filter))
		}

		if aggregateFnName != "no" {
			result["aggregate"] = aggregateFnName
			rs, err := client.QueryAggregate(nil, stmt, "record_udfs", aggregateFnName)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			select {
			case rec := <-rs.Records:
				if rec != nil {
					result["data"] = rec.Bins["SUCCESS"]
				} else {
					result["data"] = 0
				}
				break
			case err := <-rs.Errors:
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err})
					return
				}
			}
		} else {
			_, err := client.QueryObjects(nil, stmt, asResults)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			} else {
				data := make([]*models.BidRequest, 0, 10)
				cnt := 0
				for res := range asResults {
					data = append(data, res)
					cnt++
					if cnt >= LIMIT {
						result["isPartial"] = true
						break
					}
				}
				result["data"] = data
				result["count"] = cnt
			}
		}
		c.JSON(200, result)
	}
}

func GetFilteringFields() []string {
	return []string{"Timestamp", "Action", "SSP"}
}

func GetAggragateFnNames() []string {
	return []string{"Count"}
}
