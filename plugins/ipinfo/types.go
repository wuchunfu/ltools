package ipinfo

import "time"

// IPInfo IP信息结构
type IPInfo struct {
	IP          string    `json:"ip"`
	Country     string    `json:"country,omitempty"`
	CountryCode string    `json:"countryCode,omitempty"`
	Region      string    `json:"region,omitempty"`
	RegionName  string    `json:"regionName,omitempty"`
	City        string    `json:"city,omitempty"`
	ISP         string    `json:"isp,omitempty"`
	Org         string    `json:"org,omitempty"`
	Timezone    string    `json:"timezone,omitempty"`
	Lat         float64   `json:"lat,omitempty"`
	Lon         float64   `json:"lon,omitempty"`
	Query       string    `json:"query"`
	FetchedAt   time.Time `json:"fetchedAt"`
}

// LocalIPInfo 本地IP信息
type LocalIPInfo struct {
	Interface string `json:"interface"` // 网络接口名称
	IP        string `json:"ip"`        // IP地址
	IPv6      string `json:"ipv6"`      // IPv6地址（如果有）
	MAC       string `json:"mac"`       // MAC地址
}

