package weather

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	Seven_DAY_INFO_START = "<div class=\"weather_7d\">"
	Seven_DAY_INFO_END   = "</script>"
	UPDATE_TIME_FOUND    = "<input type=\"hidden\" id=\"update_time\" value=\"(.*?)\""
	DATE_CONTAINER_START = "<ul class=\"date-container\">"
	DATE_CONTAINER_END   = "</ul>"
	SUN_CONTAINER_START  = "<ul class=\"blue-container sky\">"

	SUN_CONTAINER_END = "<li class=\"drawTwo-container\">"
	DATE_NUM          = "<p class=\"date\">(.*?)</p>"
	DATE_NAME         = "<p class=\"date-info\">(.*?)</p>"

	DATE_WEATHER         = "<p class=\"weather-info\" title=\"(.*?)</p>"
	DATE_WINDY_LEVEL     = "<p class=\"wind-info\">(.*?)</p>"
	DATE_WINDY_DIRECTION = "<i class=\"wind-icon (.*?)\"></i>"

	LIVE_INDEX_START = "<div class=\"weather_shzs\">"
	LIVE_INDEX_END   = "</div>\n</div>"
	LIVE_INDEX_NAME  = "<h2>(.*?)</h2>"
	LIVE_INDEX_LEVEL = "<em>(.*?)</em>"
	LIVE_INDEX_TIPS  = "<dd>(.*?)</dd>"
	LIVE_INDEX_STARS = "(?s)<p[^>]*>(.*?)</p>"
	STAR             = "☆"
)

var ()

func (c *Weather) get7DaysWeatherInfoByCityNew(cityinfo RegionInfo, isFirst bool) (Resp *WeatherInfo, err error) {

	req, err := http.NewRequest("GET", WEATHER_SITE+strings.Replace(cityinfo.Url_, "weather", "weathern", 1), nil)
	if err != nil {
		log.Println(err)
		return
	}
	handle := &http.Client{Timeout: 10 * time.Second}
	resp, err := handle.Do(req)
	if nil != err || resp.StatusCode != http.StatusOK {
		log.Println(req.URL, resp.Status)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("read resp.body error", err)
		return
	}

	/*save to file*/
	//ioutil.WriteFile(fmt.Sprintf("weather_file/%s_%s.shtml", cityinfo.FullName_, cityinfo.Code_), body, 0644)
	//找到相关大块
	day7_start_re := regexp.MustCompile(Seven_DAY_INFO_START)
	day7_end_re := regexp.MustCompile(Seven_DAY_INFO_END)
	update_time_re := regexp.MustCompile(UPDATE_TIME_FOUND)
	date_container_start_re := regexp.MustCompile(DATE_CONTAINER_START)

	date_container_end_re := regexp.MustCompile(DATE_CONTAINER_END)
	//sun_constainer_start_re := regexp.MustCompile(SUN_CONTAINER_START)
	//sun_container_end_re := regexp.MustCompile(SUN_CONTAINER_END)
	live_index_start_re := regexp.MustCompile(LIVE_INDEX_START)
	live_index_end_re := regexp.MustCompile(LIVE_INDEX_END)
	//大块中过滤小块
	//日期
	date_num_re := regexp.MustCompile(DATE_NUM)
	date_name_re := regexp.MustCompile(DATE_NAME)
	//生活指数
	live_index_name_re := regexp.MustCompile(LIVE_INDEX_NAME)
	live_index_level_re := regexp.MustCompile(LIVE_INDEX_LEVEL)
	live_index_tips_re := regexp.MustCompile(LIVE_INDEX_TIPS)
	live_index_stars_re := regexp.MustCompile(LIVE_INDEX_STARS)
	//天气
	date_weather_re := regexp.MustCompile(DATE_WEATHER)
	date_windy_level := regexp.MustCompile(DATE_WINDY_LEVEL)
	date_windy_driection := regexp.MustCompile(DATE_WINDY_DIRECTION)
	//温度与日出日落
	day7_start_index := day7_start_re.FindStringIndex(string(body))
	day7_end_index := day7_end_re.FindStringIndex(string(body[day7_start_index[0]:]))
	weather_str := body[day7_start_index[1] : day7_start_index[0]+day7_end_index[1]]

	dayInfos := strings.Split(string(weather_str), "\r\n")
	dayInfos = strings.Split(dayInfos[2], "\n")
	var eventDay = make([]int, 0)
	var eventNight = make([]int, 0)
	var sunrise = make([]string, 0)
	var sunset = make([]string, 0)
	for i, v := range dayInfos {
		switch i {
		case 0: //eventDay
			eventDay = getTemperatureFromGroup(v)
		case 1: //eventNight
			eventNight = getTemperatureFromGroup(v)
		case 4: //sunup
			sunrise = getSunsetFromGroup(v)
		case 5: //sunset
			sunset = getSunsetFromGroup(v)
		default:
			// do nothing
		}
	}

	SevenDaysWeatherInfo := &WeatherInfo{
		Name_:     cityinfo.Name_,
		FullName_: cityinfo.FullName_,
		Code_:     cityinfo.Code_,
		Url_:      cityinfo.Url_,
		Spell_:    cityinfo.Spell_,
	}

	//查询当前信息
	GetCurrentWeatherInfo(cityinfo.Code_, fmt.Sprintf(CURRENT_INFO_API, cityinfo.Code_, time.Now().Nanosecond()), SevenDaysWeatherInfo)
	SevenDaysWeatherInfo.curGetTime_ = time.Now()

	/*parse 7days weather*/
	//更新时间
	tmpIndex := day7_start_index[1] + day7_end_index[1] + 300
	uptime := update_time_re.FindStringSubmatch(string(body[(day7_start_index[1] + day7_end_index[0]):tmpIndex]))
	SevenDaysWeatherInfo.UpdateTime_ = uptime[1]

	//7天具体日期 1500:这段数据大概1500个字符吧
	dateStr := body[tmpIndex-100 : tmpIndex+1500]
	s := date_container_start_re.FindAllIndex(dateStr, 1)
	e := date_container_end_re.FindAllIndex(dateStr, 1)
	dateStr = dateStr[s[0][0]:e[0][0]]
	dateNum := date_num_re.FindAllStringSubmatch(string(dateStr), -1)
	dateName := date_name_re.FindAllStringSubmatch(string(dateStr), -1)

	//天气、风，大概3500字符吧
	sunStr := body[tmpIndex+800 : tmpIndex+5000]
	//s = sun_constainer_start_re.FindAllIndex(sunStr, 1)
	//e = sun_container_end_re.FindAllIndex(sunStr[s[0][0]:], 1)
	//windInfos := sunStr[s[0][0]:e[0][1]]
	weatherStrs := date_weather_re.FindAllStringSubmatch(string(sunStr), -1)
	weatherWindyLevels := date_windy_level.FindAllStringSubmatch(string(sunStr), -1)
	weatherWindyDirections := date_windy_driection.FindAllStringSubmatch(string(sunStr), -1)

	//生活指数
	s = live_index_start_re.FindAllIndex(body, 1)
	e = live_index_end_re.FindAllIndex(body[s[0][0]:], 1)
	liveStr := body[s[0][0] : e[0][0]+s[0][0]]
	liveNames := live_index_name_re.FindAllStringSubmatch(string(liveStr), 8)
	liveLevels := live_index_level_re.FindAllStringSubmatch(string(liveStr), 8)
	liveTips := live_index_tips_re.FindAllStringSubmatch(string(liveStr), 8)
	liveStarS := live_index_stars_re.FindAllStringSubmatch(string(liveStr), 8)

	for i := 0; i < WEATHER_DAYS; i++ {
		var brief_weather_info = &BriefWeatherInfo{}
		brief_weather_info.Date_ = fmt.Sprintf("%s(%s)", dateNum[i+1][1], dateName[i+1][1])
		brief_weather_info.Sun_ = strings.Split(weatherStrs[i][1], ">")[1]
		brief_weather_info.Temperature_[0] = eventNight[i+1]
		brief_weather_info.Temperature_[1] = eventDay[i+1]
		brief_weather_info.Turn_.Sunrise = sunrise[i]
		brief_weather_info.Turn_.Sunset = sunset[i]
		brief_weather_info.Wind_.From_ = strings.Split(weatherWindyDirections[i+2][1], "\"")[2]
		brief_weather_info.Wind_.To_ = strings.Split(weatherWindyDirections[i+3][1], "\"")[2]
		brief_weather_info.Wind_.Level_ = weatherWindyLevels[i][1]
		//fmt.Println(dayInfo, brief_weather_info)
		SevenDaysWeatherInfo.Weather_[i] = brief_weather_info
	}
	for i := 0; i < LIVE_INDEX_INFO_COUNT; i++ {
		SevenDaysWeatherInfo.LiveIndex_[i] = &LiveIndexInfo{
			Name_:  liveNames[i][1],
			Level_: liveLevels[i][1],
			Tips:   liveTips[i][1],
			Stars_: fillStars(strings.Count(liveStarS[i][1], "active")),
		}
	}

	//设定更新时间
	SevenDaysWeatherInfo.getime_ = time.Now()

	//查询是需要获取告警信息
	locations, ok := GetLocationInfoByID(cityinfo.Code_)
	if ok {
		SevenDaysWeatherInfo.AlarmInfo_ = SevenDaysWeatherInfo.AlarmInfo_[:0]
		for _, v := range locations {
			GetAlarmDetails(ALARM_DETAILS+v.FileName, SevenDaysWeatherInfo)
		}
	}
	c.addWeatherInfoToCache(cityinfo.Code_, SevenDaysWeatherInfo)
	return SevenDaysWeatherInfo, err
}

func getTemperatureFromGroup(str string) []int {
	s := strings.Index(str, "[")
	strTmp := str[s+1 : len(str)-2]
	t := make([]int, 0)
	var st = strings.Split(strTmp, ",")
	for _, v := range st {
		v = strings.Trim(v, "\"")
		nt, _ := strconv.ParseInt(v, 10, 64)
		t = append(t, int(nt))
	}
	return t
}

func getSunsetFromGroup(str string) []string {
	s := strings.Index(str, "[")
	strTmp := str[s+1 : len(str)-2]
	t := make([]string, 0)
	var st = strings.Split(strTmp, ",")
	for _, v := range st {
		v = strings.Trim(v, "\"")
		t = append(t, v)
	}
	return t
}
func fillStars(c int) string {
	var stars string
	for i := 0; i < c; i++ {
		stars = stars + STAR
	}
	return stars
}
