package views

import (
	"github.com/gin-gonic/gin"
	"adspend/models"
	"time"
	"net/http"
	as "github.com/aerospike/aerospike-client-go"
	"strings"
	"strconv"
)

const namespace = "test"
const setName = "adspend"

const GIG int64 = 1000000000

type RecordView struct {
	Client *as.Client
	Policy *as.WritePolicy
}

func (rv RecordView) Init() {
	rv.Policy = as.NewWritePolicy(0, 0)
	rv.Policy.Timeout = 50 * time.Millisecond  // 50 millisecond timeout.
	rv.Policy.Expiration = 3600 * 24 * 60 // 60 days NB! this is time how long records will be store
}

func (rv RecordView) Post(c *gin.Context) {
	var response struct {
		Success bool
		Message string
	}
	var record models.BidRequest
	if c.BindJSON(&record) == nil {
		record.Timestamp = time.Now()
		key, _ := as.NewKey(namespace, setName, record.Timestamp.Format(time.ANSIC))

		if err := rv.Client.PutObject(rv.Policy, key, &record); err == nil {
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
}

func (rv RecordView) CreateHandler(field string, aggregateFnName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		filterBy := c.Param("filter")
		result := gin.H{"field": field, "filterBy": filterBy, "ts": time.Now().Unix() }
		data := make([]interface{}, 0, 20)
		stmt := as.NewStatement(namespace, setName)
		asResults := make(chan *models.BidRequest, 10)
		if strings.Contains(filterBy, "-") && field == "Timestamp" {
			filterValues := strings.Split(filterBy, "-")
			var tsFrom int64
			var tsTo int64
			var err error
			if filterValues[0] != "" {
				if tsFrom, err = strconv.Atoi(filterValues[0]); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err})
					return
				}
			} else {
				tsFrom = 0
			}
			if filterValues[1] != "" {
				if tsTo, err = strconv.Atoi(filterValues[1]); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err})
					return
				}
			} else {
				tsTo = 0xFFFFFFFF
			}

			stmt.Addfilter(as.NewRangeFilter(field, tsFrom * GIG, (tsTo + 1) * GIG))
		} else {
			stmt.Addfilter(as.NewEqualFilter(field, filterBy))
		}

		if aggregateFnName != "no" {
			result["aggregate"] = aggregateFnName
		}
		_, err := rv.Client.QueryObjects(nil, stmt, asResults)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		} else {
			for res := range asResults {
				pk := len(data)
				data = data[:pk + 1]
				data[pk] = res
			}
			result["data"] = data
		}
		c.JSON(200, result)
	}
}

func (rv RecordView) GetFilteringFields() []string {
	return []string{"Timestamp", "Action", "SSP"}
}

func (rv RecordView) GetAggragateFnNames() []string {
	return []string{"Count", "Max", "Min"}
}