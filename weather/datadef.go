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
	Alarm_      bool   `json:"alarm"`
	getime_     time.Time
	FullName_   string                                `json:"fullname"`
	LiveIndex_  [LIVE_INDEX_INFO_COUNT]*LiveIndexInfo `json:"liveindex"`
	Weather_    [WEATHER_DAYS]*BriefWeatherInfo       `json:"weather"`
	AlarmInfo_  AlarmDetails                          `json:"alarminfo"`
}

type CacheStats struct {
	Bytes       int64 `json:"bytes"`
	Items       int64 `json:"items"`
	Gets        int64 `json:"gets"`
	Hits        int64 `json:"hits"`
	Evictions   int64 `json:"evictions"`
	RefreshRate int64 `json:"refreshrate"`
}

//----------------------------

// Location 表示地点详细信息的结构体
type Location struct {
	Name      string `json:"name"`      // 地点名称
	FileName  string `json:"file_name"` // 文件名
	Longitude string `json:"longitude"` // 经度
	Latitude  string `json:"latitude"`  // 纬度
	Code      string `json:"code"`      // 地区代码
	Code2     string `json:"code2"`     // 地区代码2，可能是重复的地区代码
}

// AlarmInfoResp 表示整个AlarmList数据的结构体
type AlarmInfoResp struct {
	Count string     `json:"count"` // 计数
	Data  [][]string `json:"data"`  // 地点数据列表
}

// AlarmDetails	用于填充详情页面的数据
type AlarmDetails struct {
	Title    string `json:"title"`    //标题
	Details  string `json:"details"`  //上下文
	Standard string `json:"standard"` //预警标准
	Manual   string `json:"manual"`   //防御措施
}

// ---
// 定义结构体
type STEMP1 struct {
	Head         string `json:"head"`
	AlertID      string `json:"ALERTID"`
	Province     string `json:"PROVINCE"`
	City         string `json:"CITY"`
	StationName  string `json:"STATIONNAME"`
	SignalType   string `json:"SIGNALTYPE"`
	SignalLevel  string `json:"SIGNALLEVEL"`
	TypeCode     string `json:"TYPECODE"`
	LevelCode    string `json:"LEVELCODE"`
	IssueTime    string `json:"ISSUETIME"`
	IssueContent string `json:"ISSUECONTENT"`
	Underwriter  string `json:"UNDERWRITER"`
	RelieveTime  string `json:"RELIEVETIME"`
	NameEN       string `json:"NAMEEN"`
	YJTypeEN     string `json:"YJTYPE_EN"`
	YJYCEN       string `json:"YJYC_EN"`
	Time         string `json:"TIME"`
	Effect       string `json:"EFFECT"`
	MsgType      string `json:"msgType"`
	Identifier   string `json:"identifier"`
	References   string `json:"references"`
}
