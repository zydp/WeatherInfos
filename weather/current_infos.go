package weather

import (
	"encoding/json"
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
	CURRENT_INFO_API = "http://d1.weather.com.cn/sk_2d/%s.html?_=%d" //needs a timestamp, for example： 1719477278246
	HOUR_INFOS_URL   = "http://www.weather.com.cn/weather1dn/%s.shtml"
	HOUR_INFO_START  = "<div class=\"todayRight\">"
	HOUR_INFO_END    = "var hour3week="
)

func GetCurrentWeatherInfo(code, rawURL string, r *WeatherInfo) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		log.Println(err)
		return
	}
	//cheat
	req.Header.Add("Host", "d1.weather.com.cn")
	req.Header.Add("Referer", "http://www.weather.com.cn/")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0")
	handle := &http.Client{Timeout: 10 * time.Second}
	resp, err := handle.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(req.URL, err)
		return
	}
	defer resp.Body.Close()
	buf, _ := ioutil.ReadAll(resp.Body)
	data := buf[11:]
	var curinfo CurrentWeatherInfo
	json.Unmarshal(data, &curinfo)
	r.CurrentInfo.Date = curinfo.Date
	r.CurrentInfo.Time = curinfo.Time
	r.CurrentInfo.AirQuality = curinfo.AirQuality
	//r.CurrentInfo.AirQualityPM25 = curinfo.AirQualityPM25
	r.CurrentInfo.Visibility = curinfo.Visibility
	r.CurrentInfo.Pressure = curinfo.Pressure
	r.CurrentInfo.Humidity = curinfo.Humidity
	r.CurrentInfo.WindLevel = curinfo.WindSpeed
	r.CurrentInfo.WindSpeed = curinfo.WindSpeedE
	r.CurrentInfo.WindDirection = curinfo.WindDirection
	r.CurrentInfo.Temperature = curinfo.Temperature
	r.CurrentInfo.TemperatureF = curinfo.TemperatureF
	r.CurrentInfo.Weather = curinfo.Weather
	getHourInfos(fmt.Sprintf(HOUR_INFOS_URL, code), r)
}

func getHourInfos(rawURL string, r *WeatherInfo) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		log.Println(err)
		return
	}
	//cheat
	req.Header.Add("Host", "d1.weather.com.cn")
	req.Header.Add("Referer", "http://www.weather.com.cn/")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0")
	handle := &http.Client{Timeout: 10 * time.Second}
	resp, err := handle.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(req.URL, err)
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	hour_start_re := regexp.MustCompile(HOUR_INFO_START)
	hour_end_re := regexp.MustCompile(HOUR_INFO_END)

	hour_start_index := hour_start_re.FindStringIndex(string(body))
	hour_end_index := hour_end_re.FindStringIndex(string(body[hour_start_index[0]:]))
	tempInfos := body[hour_start_index[1] : hour_start_index[0]+hour_end_index[1]]
	hourInfos := strings.Split(string(strings.Split(string(tempInfos), "=")[1]), ";")[0]
	var hours = make([][]wwwHourInfos, 0)
	var rhours = make([][]HourInfos, 0)
	json.Unmarshal([]byte(hourInfos), &hours)
	for i := 0; i < 3; i++ {
		var days = make([]HourInfos, 0)
		for _, v := range hours[i] {
			temp, _ := strconv.Atoi(v.Jb)
			dl, _ := strconv.Atoi(v.Jc)
			di, _ := strconv.Atoi(v.Jd)
			var year, month, day, hour int
			year, _ = strconv.Atoi(string(v.Jf[:4]))
			month, _ = strconv.Atoi(string(v.Jf[4:6]))
			day, _ = strconv.Atoi(string(v.Jf[6:8]))
			hour, _ = strconv.Atoi(string(v.Jf[8:]))
			days = append(days, HourInfos{
				Weather: GetWeatherByCode(v.Ja), Temp: temp,
				WindDirection: GetWindDirectionByIndex(di), WindLevel: GetWindLevelByIndex(dl),
				Year: year, Month: month, Day: day, Hour: hour})
		}
		rhours = append(rhours, days)
	}
	r.HoursPredict_ = rhours
}
