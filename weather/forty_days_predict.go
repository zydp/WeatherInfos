package weather

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
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

	// 从缓存读取数据
	c.fortyMu.RLock()
	v, ok := c.fortydayslru.Get(cityinfo.Code_)
	c.fortyMu.RUnlock()
	
	if ok {
		cachedData := v.([]FortyDaysInfo)
		if len(cachedData) > 0 && !timeCheckNew(cachedData[0].updateTime_, float64(UPDATE_WEATHERINFO_GAP_MINUTES)) {
			return cachedData, nil
		}
		// 如果缓存数据过期但不为空，先返回缓存数据，同时异步更新
		if len(cachedData) > 0 {
			go func() {
				c.updateFortyDaysData(cityinfo)
			}()
			return cachedData, nil
		}
	}

	// 获取新数据
	return c.updateFortyDaysData(cityinfo)
}

func (c *Weather) updateFortyDaysData(cityinfo RegionInfo) ([]FortyDaysInfo, error) {
	tNow := time.Now()
	var rfortyInfos = make([]FortyDaysInfo, 0)
	
	//40天一般跨月了，所以请求两次
	getFortyDaysInfoImpl(tNow.Year(), int(tNow.Month()), cityinfo.Code_, &rfortyInfos)
	y, m := getNextMonth(tNow)
	getFortyDaysInfoImpl(y, m, cityinfo.Code_, &rfortyInfos)

	if len(rfortyInfos) == 0 {
		return nil, errors.New("failed to fetch weather data")
	}

	//write to cache
	c.fortyMu.Lock()
	c.fortydayslru.Add(cityinfo.Code_, rfortyInfos)
	c.fortyMu.Unlock()
	
	return rfortyInfos, nil
}

func getFortyDaysInfoImpl(year, month int, code string, r *[]FortyDaysInfo) {
	maxRetries := 3
	var err error
	var resp *http.Response
	
	for i := 0; i < maxRetries; i++ {
		req, err := http.NewRequest("GET", fmt.Sprintf(FORTY_DAYS_PREDICT_URL, year, code, year, month), nil)
		if err != nil {
			log.Println(err)
			continue
		}
		//cheat
		req.Header.Add("Host", "d1.weather.com.cn")
		req.Header.Add("Referer", "http://www.weather.com.cn/")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36 Edg/124.0.0.0")
		handle := &http.Client{Timeout: 10 * time.Second}
		resp, err = handle.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
		log.Printf("Attempt %d failed: %v\n", i+1, err)
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Second * time.Duration(i+1))
	}
	
	if err != nil || resp == nil {
		log.Printf("Failed to fetch data after %d attempts for code %s: %v\n", maxRetries, code, err)
		return
	}
	defer resp.Body.Close()
	
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read response body: %v\n", err)
		return
	}
	
	if len(buf) <= 11 {
		log.Printf("Response too short for code %s\n", code)
		return
	}
	
	var fortyInfos = make([]CalendarInfo, 36)
	err = json.Unmarshal(buf[11:], &fortyInfos)
	if err != nil {
		log.Printf("Failed to unmarshal data for code %s: %v\n", code, err)
		return
	}
	
	validDataCount := 0
	for _, v := range fortyInfos {
		if "" == v.C1 && "" == v.C2 && "" == v.W1 {
			continue
		}
		
		// 处理风级格式
		windLevel := v.Wd1
		if strings.Contains(windLevel, "<") {
			windLevel = strings.Replace(windLevel, "<", "小于", -1)
		}
		
		var tmp = FortyDaysInfo{
			Date: v.Date, Week: "周" + v.Wk, Lunar: fmt.Sprintf("%s %s", v.Nlyf, v.Nl),
			Festival: v.Yl, RFestival: v.Fe, SolarTerm: v.Jq, SubSolarTerm: v.Winter,
			WCodeOne: v.C1, WCodeTwo: v.C2,
			Weather: v.W1, Wind: windLevel,
			HTemp: v.Max, MTemp: v.Min, HMax: v.Hmax, HMin: v.Hmin, HRate: v.Hgl, HRain: v.Rainobs,
			Ripe: v.Als, Avoid: v.Alins,
			updateTime_: time.Now(),
		}
		if "" == tmp.Weather && "" != tmp.WCodeOne {
			tmp.Weather = GetWeatherByCode(tmp.WCodeOne)
		}
		*r = append(*r, tmp)
		validDataCount++
	}
	
	if validDataCount == 0 {
		log.Printf("No valid data found for code %s\n", code)
	} else {
		log.Printf("Successfully processed %d records for code %s\n", validDataCount, code)
	}
}
