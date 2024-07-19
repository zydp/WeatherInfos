package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	FORTY_DAYS_PREDICT_URL = "https://d1.weather.com.cn/calendarFromMon/%04d/%s_%04d%02d.html"
)

func getNextMonth(t time.Time) (int, int) {
	nextMonth := t.AddDate(0, 1, 0)
	return nextMonth.Year(), int(nextMonth.Month())
}

func (c *Weather) GetFortyDaysInfoWeatherCom(province, district, city string) (r []FortyDaysInfo, err error) {

	if "" == province {
		return nil, errors.New("bad parameter")
	}
	if "" == district {
		district = province
		city = province
	}
	if "" == city {
		city = district
	}
	c.regionMu.RLock()
	var cityinfo RegionInfo
	if province, h1 := c.treeRegion.Regions[province]; h1 { /*TODO: make it recursion*/
		if dist, h2 := province.Regions[district]; h2 {
			if city, h3 := dist.Regions[city]; h3 {
				cityinfo = city.RegionInfo
			}
		}
	}
	c.regionMu.RUnlock()
	if "" == cityinfo.Code_ || "" == cityinfo.Url_ {
		return nil, errors.New("not found this city")
	}
	//-------------------------------

	{ //read from cache
		c.fortyMu.RLock()
		v, ok := c.fortydayslru.Get(cityinfo.Code_)
		c.fortyMu.RUnlock()
		if ok && !timeCheckNew(v.([]FortyDaysInfo)[0].updateTime_, float64(UPDATE_WEATHERINFO_GAP_MINUTES)) {
			return v.([]FortyDaysInfo), nil
		}
	}
	tNow := time.Now()
	var rfortyInfos = make([]FortyDaysInfo, 0)
	//40天一般跨月了，所以请求两次
	getFortyDaysInfoImpl(tNow.Year(), int(tNow.Month()), cityinfo.Code_, &rfortyInfos)
	y, m := getNextMonth(tNow)
	getFortyDaysInfoImpl(y, m, cityinfo.Code_, &rfortyInfos)

	//write to cache
	c.fortyMu.Lock()
	c.fortydayslru.Add(cityinfo.Code_, rfortyInfos)
	c.fortyMu.Unlock()
	return rfortyInfos, nil
}

func getFortyDaysInfoImpl(year, month int, code string, r *[]FortyDaysInfo) {
	req, err := http.NewRequest("GET", fmt.Sprintf(FORTY_DAYS_PREDICT_URL, year, code, year, month), nil)
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
	var fortyInfos = make([]CalendarInfo, 36)

	err = json.Unmarshal(buf[11:], &fortyInfos)
	for _, v := range fortyInfos {
		if "" == v.C1 && "" == v.C2 && "" == v.W1 {
			continue
		}
		var tmp = FortyDaysInfo{
			Date: v.Date, Week: "周" + v.Wk, Lunar: fmt.Sprintf("%s %s", v.Nlyf, v.Nl),
			Festival: v.Yl, RFestival: v.Fe, SolarTerm: v.Jq, SubSolarTerm: v.Winter,
			WCodeOne: v.C1, WCodeTwo: v.C2,
			Weather: v.W1, Wind: v.Wd1,
			HTemp: v.Max, MTemp: v.Min, HMax: v.Hmax, HMin: v.Hmin, HRate: v.Hgl, HRain: v.Rainobs,
			Ripe: v.Als, Avoid: v.Alins,
			updateTime_: time.Now(),
		}
		if "" == tmp.Weather && "" != tmp.WCodeOne {
			tmp.Weather = GetWeatherByCode(tmp.WCodeOne)
		}
		*r = append(*r, tmp)
	}
}
