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
	ServerTime_ string `json:"servertime"`
	getime_     time.Time
	curGetTime_ time.Time
	FullName_   string                                `json:"fullname"`
	CurrentInfo BriefCurrentWeatherInfo               `json:"nowinfo"`
	LiveIndex_  [LIVE_INDEX_INFO_COUNT]*LiveIndexInfo `json:"liveindex"`
	Weather_    [WEATHER_DAYS]*BriefWeatherInfo       `json:"weather"`
	AlarmInfo_  []AlarmDetails                        `json:"alarminfo"`
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
	Title       string `json:"title"`       //标题
	Details     string `json:"details"`     //上下文
	Standard    string `json:"standard"`    //预警标准
	Manual      string `json:"manual"`      //防御措施
	TypeCode    string `json:"typecode"`    //类型
	LevelCode   string `json:"levelcode"`   //级别
	SignalType  string `json:"signaltype"`  //类型
	SignalLevel string `json:"signallevel"` //级别
	IssueTime   string `json:"issuetime"`   //时间
	//Color       string `json:"color"`       //颜色
	//PicUri      string `json:"picuri"`      //uri
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

// / 当前天气信息
type CurrentWeatherInfo struct {
	NameEn         string `json:"nameen"`
	CityName       string `json:"cityname"`
	City           string `json:"city"`
	Temperature    string `json:"temp"`
	TemperatureF   string `json:"tempf"`
	WindDirection  string `json:"WD"`
	WindDirectionE string `json:"wde"` // 假设是英文缩写
	WindSpeed      string `json:"WS"`
	WindSpeedE     string `json:"wse"` // 假设是英文单位
	Humidity       string `json:"SD"`  // 假设是百分比
	HumidityE      string `json:"sd"`  // 假设是百分比
	Pressure       string `json:"qy"`
	Visibility     string `json:"njd"`
	Time           string `json:"time"`
	Rain           string `json:"rain"`
	Rain24h        string `json:"rain24h"`
	AirQuality     string `json:"aqi"`
	AirQualityPM25 string `json:"aqi_pm25"`
	Weather        string `json:"weather"`
	WeatherE       string `json:"weathere"`
	WeatherCode    string `json:"weathercode"`
	LimitNumber    string `json:"limitnumber"`
	Date           string `json:"date"`
}

type BriefCurrentWeatherInfo struct {
	Temperature   string `json:"temp"`
	TemperatureF  string `json:"tempf"`
	WindDirection string `json:"windirection"`
	WindLevel     string `json:"windlevel"`
	WindSpeed     string `json:"windspeed"`
	Humidity      string `json:"humidity"`
	Pressure      string `json:"pressure"`
	Visibility    string `json:"visibility"`
	Time          string `json:"time"`
	AirQuality    string `json:"aqi"`
	//AirQualityPM25 string `json:"aqi_pm25"`
	Weather string `json:"weather"`
	Date    string `json:"date"`
}
