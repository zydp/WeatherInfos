package weather

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	ALARM_LIST_API = "http://product.weather.com.cn/alarm/grepalarm_cn.php?_=" //needs a timestamp, for example： 1719477278246
	//"http://www.weather.com.cn/alarm/newalarmcontent.shtml?file="
	ALARM_DETAILS   = "http://product.weather.com.cn/alarm/webdata/"
	ALARM_FORM_INFO = "http://www.weather.com.cn/data/alarminfo/%s?_=%d"
)

// 备用 https://d1.weather.com.cn/dingzhi/101020100.html?_=1721292263961
var (
	mu         sync.RWMutex
	alarmInfos map[string][]Location
)

func init() {
	alarmInfos = make(map[string][]Location)
}

// GetLocationInfoByID foreign
func GetLocationInfoByID(cityCode string) (details []Location, ok bool) {
	mu.Lock()
	defer mu.Unlock()
	c, ok := alarmInfos[cityCode] //city : cityCode for example is 101220101
	if !ok {
		d, ok := alarmInfos[cityCode[:len(cityCode)-2]] //district 1012201
		if !ok {
			p, ok := alarmInfos[cityCode[:len(cityCode)-4]] //province 1012201
			if !ok {
				return nil, false
			}
			return p, ok
		}
		return d, ok
	}
	return c, ok
}

// CheckAlarmListFromWeatherCom 定时轮询告警列表
func CheckAlarmListFromWeatherCom() {
	for {
		req, err := http.NewRequest("GET", ALARM_LIST_API+fmt.Sprintf("%d", time.Now().Nanosecond()), nil)
		if err != nil {
			log.Println(err)
			return
		}
		handle := &http.Client{Timeout: 10 * time.Second}
		resp, err := handle.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Println(req.URL, err)
			return
		}
		defer resp.Body.Close()
		buf, _ := ioutil.ReadAll(resp.Body)
		var alarmInfoResp AlarmInfoResp

		err = json.Unmarshal(buf[14:len(buf)-1], &alarmInfoResp)
		{
			mu.Lock()
			alarmInfos = make(map[string][]Location) //每次清空map，以免数据重复
			for _, v := range alarmInfoResp.Data {
				var id string = strings.Split(v[1], "-")[0]
				alarmInfos[id] = append(alarmInfos[id], Location{Name: v[0], FileName: v[1], Longitude: v[2], Latitude: v[3], Code: v[4], Code2: v[5]})
			}
			mu.Unlock()
		}

		<-time.Tick(time.Duration(UPDATE_WEATHERINFO_GAP_MINUTES) * time.Minute)
	}
}

// 提取表格中的文本信息
func extractTableText(n *html.Node) (string, error) {
	if n.Type == html.ElementNode && n.Data == "tr" {
		var texts []string
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if c.Type == html.TextNode {
				texts = append(texts, c.Data)
			}
		}
		if len(texts) >= 2 {
			return texts[1], nil // 假设标准和防御指南在第二列
		}
	}
	return "", fmt.Errorf("no table text found")
}

func GetAlarmDetails(url string, r *WeatherInfo) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return
	}
	handle := &http.Client{Timeout: 10 * time.Second}
	resp, err := handle.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(req.URL, err)
		return
	}
	defer resp.Body.Close()

	buf, _ := ioutil.ReadAll(resp.Body)
	var st1 STEMP1
	err = json.Unmarshal(buf[14:], &st1)
	if nil != err {
		fmt.Println(err.Error())
	}

	fileName, _ := getFileNameFromURL(url)
	r.Alarm_ = true
	var ainfo AlarmDetails
	//r.AlarmInfo_.Title = st1.Head
	ainfo.Details = st1.IssueContent
	ainfo.IssueTime = st1.IssueTime
	ainfo.TypeCode = st1.TypeCode
	ainfo.LevelCode = st1.LevelCode
	ainfo.SignalType = st1.SignalType
	ainfo.SignalLevel = st1.SignalLevel
	//ainfo.Color = st1.YJYCEN
	//ainfo.PicUri = fmt.Sprintf("http://www.weather.com.cn/m2/i/about/alarmpic/%s%s.gif", ainfo.TypeCode, ainfo.LevelCode)
	getAlarmFormINfo(fmt.Sprintf(ALARM_FORM_INFO, fileName, time.Now().Nanosecond()), &ainfo)
	r.AlarmInfo_ = append(r.AlarmInfo_, ainfo)
}

func getAlarmFormINfo(rawURL string, details *AlarmDetails) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		log.Println(err)
		return
	}
	handle := &http.Client{Timeout: 10 * time.Second}
	resp, err := handle.Do(req)
	if err != nil || resp.StatusCode != http.StatusOK {
		log.Println(req.URL, err)
		return
	}
	defer resp.Body.Close()
	buf, _ := ioutil.ReadAll(resp.Body)
	data := strings.Split(string(buf[14:len(buf)-1]), ",")
	details.Title = data[1][1 : len(data[1])-1]
	details.Standard = strings.Replace(data[2], "\"", "", -1)
	details.Manual = strings.Replace(data[3], "<br>", "", -1)
	details.Manual = strings.Replace(details.Manual, "\"", "", -1)
}

func getFileNameFromURL(rawURL string) (string, error) {
	//re := regexp.MustCompile(`\d+-\d+-\d+-\d+\.html`)
	//match := re.FindString(rawURL)
	//if match == "" {
	//	return "", fmt.Errorf("no match found")
	//}
	//return match, nil
	parts := strings.Split(rawURL, "-")
	return parts[len(parts)-1], nil
}

// --------------------
