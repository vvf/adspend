package models

import (
	"time"
)

type BidRequest struct {
	Domain string `json:"domain"`
	PageURL string `json:"pageURL"`
	DomainTopic  string `json:"domainTopic"`
	PageTopic  string `json:"pageTopic"`
	BannerSize  string `json:"bannerSize"`
	DeviceType  string `json:"deviceType"`
	Country  string `json:"country"`
	City  string `json:"city"`
	OS string `json:"os"`
	UserAgent  string `json:"userAgent"`
	//UserId  string `json:"userId"`
	//UserCategory  string `json:"userCategory"`
	//DomainMinFloor  string `json:"domainMinFloor"`
	//SSP  string `json:"ssp"`
	//TrafficType  string `json:"trafficType"`
	//MobileOperator  string `json:"mobileOperator"`
	//Coverage int32 `json:"coverage",omitempty`
	Timestamp time.Time
}