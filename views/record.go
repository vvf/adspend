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
	var record models.BidRequest
	if err := c.BindJSON(&record); err == nil {
		record.Timestamp = time.Now()
		key, _ := as.NewKey(namespace, setName, record.Timestamp.Format(time.ANSIC))

		if err := client.PutObject(policy, key, &record); err == nil {
			c.JSON(http.StatusOK, gin.H{"error": nil})
		} else {
			c.JSON(http.StatusOK, gin.H{"error": err})
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"error": err})
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

type ConvertToResultFn (func(rec *as.Record, result gin.H))

func convertFromMap(rec *as.Record, result gin.H) {
	r1 := rec.Bins["SUCCESS"].(map[interface{}]interface{})
	if r1 == nil {
		result["data"] = nil
		result["count"] = 0
	} else {
		asResults := make(map[string]interface{}, len(r1))
		cnt := 0
		for v, c := range r1 {
			if v == nil {
				asResults["[null]"] = c
				result["warning"] = "Is field name valid?"
			} else {
				asResults[v.(string)] = c
			}
			cnt ++
			if cnt >= LIMIT {
				result["isPartial"] = true
				break
			}
		}
		result["data"] = asResults
		result["count"] = cnt
	}
}
func getConvertValueToAggregateFnName(aggregateFnName string) ConvertToResultFn {
	return func(rec *as.Record, result gin.H) {
		result[aggregateFnName] = rec.Bins["SUCCESS"]
		result["data"] = []interface{}{rec.Bins["SUCCESS"]}
		result["count"] = 1
	}
}

func ValuesOf(c *gin.Context) {
	field := c.Param("field")
	result := gin.H{"field": field, "ts": time.Now().Unix(), "error": nil }
	stmt := as.NewStatement(namespace, setName)
	rs, err := client.QueryAggregate(nil, stmt, "record_udfs", "ValuesOf", field)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	result[field] = nil
	select {
	case rec := <-rs.Records:
		if rec != nil {
			convertFromMap(rec, result)
		}
		break
	case err := <-rs.Errors:
		if err != nil {
			c.JSON(200, gin.H{"error": err})
			return
		}
	}
	c.JSON(http.StatusOK, result)

}
func CreateHandler(field string, aggregateFnName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := c.Param("filter")
		result := gin.H{"field": field, "filterValue": filter, "ts": time.Now().Unix(), "error": nil }
		stmt := as.NewStatement(namespace, setName)
		if strings.Contains(filter, "-") && field == "Timestamp" {
			if err := addRangeFilter(stmt, field, filter); err != nil {
				c.JSON(http.StatusOK, gin.H{"error": err})
				return
			}
		} else {
			stmt.Addfilter(as.NewEqualFilter(field, filter))
		}

		if aggregateFnName != "no" {
			result["aggregate"] = aggregateFnName
			var param string = ""
			if !!aggregatesFnConfig[aggregateFnName].HasParam {
				param = c.Param("param")
			}
			rs, err := client.QueryAggregate(nil, stmt, "record_udfs", aggregateFnName, param)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			result[aggregateFnName] = 0
			select {
			case rec := <-rs.Records:
				if rec != nil {
					convertFn := aggregatesFnConfig[aggregateFnName].convertFn
					if convertFn == nil {
						convertFn = getConvertValueToAggregateFnName(aggregateFnName)
					}
					convertFn(rec, result)
				}
				break
			case err := <-rs.Errors:
				if err != nil {
					c.JSON(200, gin.H{"error": err})
					return
				}
			}
		} else {
			asResults := make(chan *models.BidRequest, 10)
			_, err := client.QueryObjects(nil, stmt, asResults)
			if err != nil {
				c.JSON(200, gin.H{"error": err})
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

type AggregateConfig struct {
	HasParam  bool
	convertFn ConvertToResultFn
}

var aggregatesFnConfig map[string]AggregateConfig

func init() {
	aggregatesFnConfig = map[string]AggregateConfig{
		"Count": {
			HasParam:false,
			convertFn: getConvertValueToAggregateFnName("Count"),
		},
		"ValuesOf": {
			HasParam: true,
			convertFn: convertFromMap,
		},
	}
}

func GetAggregateFnNames() map[string]AggregateConfig {
	return aggregatesFnConfig
}
