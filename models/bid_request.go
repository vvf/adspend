package models

import (
	"time"
)

type BidRequest struct {
	UserId  string `json:"userId"`
	SSP  string `json:"ssp"`
	Timestamp time.Time `json:"timestamp"`
	Action string `json:"action"`
	Host string `json:"host"`
       	Path string `json:"path"`
       	Query string `json:"query"`
       	Ip string `json:"ip"`
       	Ua string `json:"ua"`
       	SeatID string `json:"seatID"`
       	CreativeSize string `json:"creativeSize"`
       	CreativeType string `json:"creativeType"`
       	Geo string `json:"geo"`
}