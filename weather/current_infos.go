package weather

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	CURRENT_INFO_API = "http://d1.weather.com.cn/sk_2d/%s.html?_=%d" //needs a timestamp, for example： 1719477278246
)

func GetCurrentWeatherInfo(rawURL string, r *WeatherInfo) {
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

}
