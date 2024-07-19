package weather

import (
	"time"
)

const (
	LIVE_INDEX_INFO_COUNT = 6
	WEATHER_DAYS          = 7
)

var (
	WEATHER_DESC = map[string]string{
		"00": "晴", "01": "多云", "02": "阴",
		"03": "阵雨", "04": "雷阵雨", "05": "雷阵雨伴有冰雹",
		"06": "雨夹雪", "07": "小雨", "08": "中雨",
		"09": "大雨", "10": "暴雨", "11": "大暴雨", "12": "特大暴雨",
		"13": "阵雪", "14": "小雪", "15": "中雪", "16": "大雪", "17": "暴雪",
		"18": "雾", "19": "冻雨", "20": "沙尘暴",
		"21": "小到中雨", "22": "中到大雨", "23": "大到暴雨", "24": "暴雨到大暴雨", "25": "大暴雨到特大暴雨",
		"26": "小到中雪", "27": "中到大雪", "28": "大到暴雪",
		"29": "浮尘", "30": "扬沙", "31": "强沙尘暴",
		"53": "霾", "99": "无",
		"32": "浓雾", "49": "强浓雾", "54": "中度霾", "55": "重度霾", "56": "严重霾", "57": "大雾", "58": "特强浓雾",
		"97": "雨", "98": "雪",
		"301": "雨", "302": "雪",
	}
	WIND_DIRECTION = []string{"无持续风向", "东北风", "东风", "东南风", "南风", "西南风", "西风", "西北风", "北风", "旋转风"}
	WIND_LEVEL     = []string{"<3级", "3-4级", "4-5级", "5-6级", "6-7级", "7-8级", "8-9级", "9-10级", "10-11级", "11-12级"}
)

func GetWeatherByCode(c string) string {
	var name string = ""
	name, _ = WEATHER_DESC[c]
	return name
}
func GetWindDirectionByIndex(i int) string {
	if i > 9 {
		return ""
	}
	return WIND_DIRECTION[i]
}
func GetWindLevelByIndex(i int) string {
	if i > 9 {
		return ""
	}
	return WIND_LEVEL[i]
}

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
type Turn struct {
	Sunrise string `json:"sunrise"`
	Sunset  string `json:"sunset"`
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
	Turn_        Turn   `json:"turn"`
}

type LiveIndexInfo struct {
	Name_  string `json:"name"`
	Level_ string `json:"level"`
	Stars_ string `json:"stars"`
	Tips   string `json:"tips"`
}

type WeatherInfo struct {
	Code_         string `json:"-"`
	Url_          string `json:"-"`
	Name_         string `json:"name"`
	Spell_        string `json:"spell"`
	UpdateTime_   string `json:"updatetime"`
	Alarm_        bool   `json:"alarm"`
	ServerTime_   string `json:"servertime"`
	Lunar_        string `json:"lunar"`
	getime_       time.Time
	curGetTime_   time.Time
	FullName_     string                                `json:"fullname"`
	CurrentInfo   BriefCurrentWeatherInfo               `json:"nowinfo"`
	HoursPredict_ [][]HourInfos                         `json:"hours"`
	LiveIndex_    [LIVE_INDEX_INFO_COUNT]*LiveIndexInfo `json:"liveindex"`
	Weather_      [WEATHER_DAYS]*BriefWeatherInfo       `json:"weather"`
	AlarmInfo_    []AlarmDetails                        `json:"alarminfo"`
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

type wwwHourInfos struct {
	Ja string `json:"ja"`
	Jb string `json:"jb"`
	Jc string `json:"jc"`
	Jd string `json:"jd"`
	Je string `json:"je"`
	Jf string `json:"jf"`
}

type HourInfos struct {
	Weather       string `json:"weather"`
	Temp          int    `json:"temp"`
	WindDirection string `json:"windDirection"`
	WindLevel     string `json:"windLevel"`
	Year          int    `json:"year"`
	Month         int    `json:"month"`
	Day           int    `json:"day"`
	Hour          int    `json:"hour"`
}

// forty days weather infos
type CalendarInfo struct {
	Alins   string `json:"alins"`   // 宜忌
	Als     string `json:"als"`     // 宜事项列表
	Blue    string `json:"blue"`    // 蓝色
	C1      string `json:"c1"`      // C1
	C2      string `json:"c2"`      // C2
	Cla     string `json:"cla"`     // 类别
	Date    string `json:"date"`    // 日期
	Des     string `json:"des"`     // 描述
	Fe      string `json:"fe"`      // FE
	Hgl     string `json:"hgl"`     // 黄道吉日百分比
	Hmax    string `json:"hmax"`    // 最高
	Hmin    string `json:"hmin"`    // 最低
	Hol     string `json:"hol"`     // 假日
	Insuit  string `json:"insuit"`  // 忌事项
	Jq      string `json:"jq"`      // 吉日
	Max     string `json:"max"`     // 最大
	Maxobs  string `json:"maxobs"`  // 最大观测
	Min     string `json:"min"`     // 最小
	Minobs  string `json:"minobs"`  // 最小观测
	Nl      string `json:"nl"`      // 农历日期
	Nlyf    string `json:"nlyf"`    // 农历月份
	R       string `json:"r"`       // R
	Rainobs string `json:"rainobs"` // 降雨观测
	Suit    string `json:"suit"`    // 宜事项
	T1      string `json:"t1"`      // T1
	T1t     string `json:"t1t"`     // T1时间
	T2      string `json:"t2"`      // T2
	T3      string `json:"t3"`      // T3
	T3t     string `json:"t3t"`     // T3时间
	Time    string `json:"time"`    // 时间
	Today   string `json:"today"`   // 今天
	Update  string `json:"update"`  // 更新
	W1      string `json:"w1"`      // W1
	Wd1     string `json:"wd1"`     // Wd1
	Winter  string `json:"winter"`  // 冬季
	Wk      string `json:"wk"`      // 星期
	Wor     string `json:"wor"`     // 工作日
	Ws1     string `json:"ws1"`     // Ws1
	Yl      string `json:"yl"`      // 节日
}

type FortyDaysInfo struct {
	Date         string `json:"date"`         //日期
	Week         string `json:"week"`         //星期
	Ripe         string `json:"ripe"`         //宜
	Avoid        string `json:"avoid"`        //忌
	Lunar        string `json:"lunar"`        //农历
	SolarTerm    string `json:"solarTerm"`    //节气
	SubSolarTerm string `json:"subSolarTerm"` //节气进度
	Festival     string `json:"festival"`     //节日
	RFestival    string `json:"rfestival"`    //节日2
	Weather      string `json:"weather"`      //天气
	WCodeOne     string `json:"wcodeOne"`     //天气代码1
	WCodeTwo     string `json:"wcodeTwo"`     //天气代码2
	Wind         string `json:"wind"`         //风向及风级
	HTemp        string `json:"htemp"`        //最高温
	MTemp        string `json:"mtemp"`        //最低温
	HMax         string `json:"hmax"`         //历史最高温
	HMin         string `json:"hmin"`         //历史最低温
	HRate        string `json:"hrate"`        //历史降雨概率
	HRain        string `json:"hrain"`        //历史降雨量
	updateTime_  time.Time
}
