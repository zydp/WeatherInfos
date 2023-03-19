package weather

import (
	"time"
)

const (
	LIVE_INDEX_INFO_COUNT = 6
	WEATHER_DAYS          = 7
)

type RegionInfo struct {
	Url_      string `json:"-"`
	Code_     string `json:"-"`
	Name_     string `json:"name"`
	FullName_ string `json:"fullname"`
	Spell_    string `json:"spell"`
}

/*
*  Nesting
*  Taiwan->Taibei->Taoyuan
 */
type TreeRegionInfo struct {
	RegionInfo
	Regions map[string]*TreeRegionInfo
}

type Wind struct {
	From_  string `json:"from"`
	To_    string `json:"to"`
	Level_ string `json:"level"`
}

type BriefWeatherInfo struct {
	Date_        string `json:"date"`
	Sun_         string `json:"sun"`
	Temperature_ [2]int `json:"temperature"`
	Wind_        Wind   `json:"wind"`
}

type LiveIndexInfo struct {
	Name_  string `json:"name"`
	Level_ string `json:"level"`
	Tips   string `json:"tips"`
}

type WeatherInfo struct {
	Code_       string `json:"-"`
	Url_        string `json:"-"`
	Name_       string `json:"name"`
	Spell_      string `json:"spell"`
	UpdateTime_ string `json:"updatetime"`
	getime_     time.Time
	FullName_   string                                `json:"fullname"`
	LiveIndex_  [LIVE_INDEX_INFO_COUNT]*LiveIndexInfo `json:"liveindex"`
	Weather_    [WEATHER_DAYS]*BriefWeatherInfo       `json:"weather"`
}

type CacheStats struct {
	Bytes     int64
	Items     int64
	Gets      int64
	Hits      int64
	Evictions int64
}
