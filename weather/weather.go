package weather

import (
	"WeatherInfos/lrucache"
	"bufio"
	"encoding/gob"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

func init() {

}

const (
	WEATHER_SITE               = "http://www.weather.com.cn"
	REGION_SITE                = "http://www.weather.com.cn/textFC/hb.shtml"
	TEMP_FILE_PATH             = "./temp"
	REGION_CACHE_FILE          = ".region_data.gob"
	REGEXP_GET_REGION_URL_INFO = "/textFC/[a-z-]{3,20}\\W\\w{5}\\W{2}\\w{6}\\W{2}\\w{6}\\W{2}[\u4e00-\u9fa5]+\\W{2}\\w>"
	REGEXP_GET_REGION_URL      = "/textFC/[a-z-]{3,20}\\W\\w{5}"
	REGEXP_GET_WORD            = "[\u4e00-\u9fa5]+"
	REGEXP_GET_CITY_URL_INFO   = "/weather/[0-9]{6,12}\\W\\w{5}\\W{2}\\w{6}\\W{2}\\w{6}\\W{2}[\u4e00-\u9fa5]+\\W{2}\\w>"
	REGEXP_GET_CITY_URL        = "/weather/[0-9]{6,12}\\W\\w{5}"
	REGEXP_GET_CITY_CODE       = "[0-9]{6,12}"
	REGEXP_GET_CITY_START      = "<div class=\"conMidtab3\">"
	REGEXP_GET_CITY_END        = "</div>"
	DISCARD_INFO_FIELD         = "详情"
	REGEXP_WEATHER_DAY7_START  = "<ul class=\"t clearfix\">"
	REGEXP_WEATHER_END         = "</ul>"
	REGEXP_GET_WEATHER_NUM     = "[0-9-]+"
	REGEXP_LIVE_INDEX_START    = "<ul class=\"clearfix\">"
	REGEXT_GET_SPORT_STAR      = "class=\"star\""
	DAY_INFO_SPLIT_SEP         = "</li>"
)

const (
	DEFAULT_LIMIT_SIZE             = 34
	UPDATE_WEATHERINFO_GAP_MINUTES = 60*4 + 30
)

type Weather struct {
	weatherMu, regionMu, entryCacheMu sync.RWMutex
	nbytes                            int64
	weatherlru                        *lrucache.Cache
	entryCache                        *lrucache.Cache
	nhit, nget                        int64
	nevict                            int64
	treeRegion                        *TreeRegionInfo
	inited                            bool
}

func New(maxEntries int) *Weather {
	return &Weather{
		weatherlru: lrucache.New(maxEntries),
		entryCache: lrucache.New(0),
		treeRegion: &TreeRegionInfo{Regions: make(map[string]*TreeRegionInfo)},
	}
}

func (c *Weather) InitRegionTree() (err error) {
	c.regionMu.Lock()
	defer c.regionMu.Unlock()
	if err = c.loadRegionData(REGION_CACHE_FILE); err != nil {
		log.Println("load region info from file failed,ready to get")
		var wg sync.WaitGroup
		req, err := http.NewRequest("GET", REGION_SITE, nil)
		if err != nil {
			log.Println(err)
			return err
		}
		//req.Header.Add("Content-Type", "application/json")
		handle := &http.Client{Timeout: 10 * time.Second}
		resp, err := handle.Do(req)
		if err != nil || resp.StatusCode != http.StatusOK {
			log.Println(req.URL, err)
			return err
		}

		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("read resp.body error", req.URL, err)
			return err
		}

		re := regexp.MustCompile(REGEXP_GET_REGION_URL_INFO) //取出连接
		regexpResult := re.FindAllString(string(body), 100)
		urlfind := regexp.MustCompile(REGEXP_GET_REGION_URL)
		namefind := regexp.MustCompile(REGEXP_GET_WORD)

		chCache := make(chan *TreeRegionInfo, 100)
		for _, strUrl := range regexpResult {
			wg.Add(1)
			go func(url, name string, cache chan<- *TreeRegionInfo) {
				defer wg.Done()
				regionInfo := &TreeRegionInfo{
					RegionInfo: RegionInfo{
						Name_: name, Url_: url,
					},
					Regions: make(map[string]*TreeRegionInfo),
				}
				c.parseCityOrCountyInfo(regionInfo)
				cache <- regionInfo
			}(urlfind.FindString(strUrl), namefind.FindString(strUrl), chCache)
		}
		wg.Wait()
		close(chCache)
		for {
			info := <-chCache
			if nil == info {
				break
			}
			c.treeRegion.Regions[info.Name_] = info
		}
		if err := c.saveRegionData(REGION_CACHE_FILE); nil != err {
			log.Println(err)
		}
	}
	return
}

func (c *Weather) parseCityOrCountyInfo(info *TreeRegionInfo) {
	req, err := http.NewRequest("GET", WEATHER_SITE+info.Url_, nil)
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

	infoExtraction := regexp.MustCompile(REGEXP_GET_CITY_URL_INFO)
	start_re := regexp.MustCompile(REGEXP_GET_CITY_START)
	end_re := regexp.MustCompile(REGEXP_GET_CITY_END)

	urlfind := regexp.MustCompile(REGEXP_GET_CITY_URL)
	namefind := regexp.MustCompile(REGEXP_GET_WORD)
	codefind := regexp.MustCompile(REGEXP_GET_CITY_CODE)
	start_index := start_re.FindAllStringIndex(string(buf), -1)

	for _, index := range start_index {
		end_index := end_re.FindStringIndex(string(buf[index[1]:]))
		sinfo := string(buf[index[0] : index[1]+end_index[1]])
		counties := infoExtraction.FindAllString(sinfo, -1)
		cityName := namefind.FindString(sinfo)
		if _, had := info.Regions[cityName]; had {
			break /*repeat*/
		}
		regionInfo := &TreeRegionInfo{
			RegionInfo: RegionInfo{Name_: cityName, FullName_: fmt.Sprintf("%s_%s", info.Name_, cityName)},
			Regions:    make(map[string]*TreeRegionInfo),
		}
		for _, cinfo := range counties {
			countyName := namefind.FindString(cinfo)
			if DISCARD_INFO_FIELD == countyName { /*skip the description*/
				continue
			}
			county := &TreeRegionInfo{
				RegionInfo: RegionInfo{
					Url_:      urlfind.FindString(cinfo),
					Code_:     codefind.FindString(cinfo),
					Name_:     countyName,
					FullName_: fmt.Sprintf("%s_%s_%s", info.Name_, cityName, countyName),
				},
			}
			regionInfo.Regions[county.Name_] = county
		}
		info.Regions[cityName] = regionInfo
	}
	if cinfos, ok := info.Regions[info.Name_]; ok {
		if county_info, had := cinfos.Regions[info.Name_]; had {
			info.Code_ = county_info.Code_
		}
	}
}

func (c *Weather) ShowCityWeather(province, district, city string) (Resp *WeatherInfo, err error) {
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

	resp, has := c.getWeatherInfoForCache(cityinfo.Code_)
	if has && !timeCheck(resp.getime_) {
		return resp, nil
	}
	if has {
		log.Printf("last update time：%s  %s\n", resp.FullName_, resp.getime_.Format(time.RFC3339))
	}
	return c.get7DaysWeatherInfoByCity(cityinfo, has)
}

func timeCheck(dataTime time.Time) (ok bool) {
	dur := time.Now().Sub(dataTime)
	return dur.Minutes() >= UPDATE_WEATHERINFO_GAP_MINUTES
}

func (c *Weather) get7DaysWeatherInfoByCity(cityinfo RegionInfo, isFirst bool) (Resp *WeatherInfo, err error) {
	req, err := http.NewRequest("GET", WEATHER_SITE+cityinfo.Url_, nil)
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

	day7_start_re := regexp.MustCompile(REGEXP_WEATHER_DAY7_START)
	live_index_re := regexp.MustCompile(REGEXP_LIVE_INDEX_START)
	sport_star_re := regexp.MustCompile(REGEXT_GET_SPORT_STAR)
	wordfind_re := regexp.MustCompile(REGEXP_GET_WORD)
	numfind_re := regexp.MustCompile(REGEXP_GET_WEATHER_NUM)
	end_re := regexp.MustCompile(REGEXP_WEATHER_END)
	//airOn_re := regexp.MustCompile("过去24小时AQI最高值: \\d+")

	day7_start_index := day7_start_re.FindStringIndex(string(body))
	day7_end_index := end_re.FindStringIndex(string(body[day7_start_index[0]:]))
	weather_str := body[day7_start_index[1] : day7_start_index[0]+day7_end_index[1]]
	dayInfos := strings.Split(string(weather_str), DAY_INFO_SPLIT_SEP)

	SevenDaysWeatherInfo := &WeatherInfo{
		Name_:     cityinfo.Name_,
		FullName_: cityinfo.FullName_,
		Code_:     cityinfo.Code_,
		Url_:      cityinfo.Url_,
	}
	/*parse 7days weather*/
	uptime := numfind_re.FindAllString(string(body[day7_start_index[0]-30:day7_start_index[0]]), 2)
	SevenDaysWeatherInfo.UpdateTime_ = fmt.Sprintf("%s:%s", uptime[0], uptime[1])
	var localWeatherIndex int = 0
	for _, dayInfo := range dayInfos {
		var brief_weather_info = &BriefWeatherInfo{}
		reader := bufio.NewReader(strings.NewReader(dayInfo))
		for line := 0; ; line++ {
			value, _, err := reader.ReadLine()
			if err != nil {
				break
			}
			switch line {
			case 2: //date
				brief_weather_info.Date_ = string(value[4 : len(value)-5])
			case 5: //sun
				brief_weather_info.Sun_ = wordfind_re.FindString(string(value))
			case 7: //temperature
				temp := numfind_re.FindAllString(string(value), 2)
				if len(temp) >= 2 {
					brief_weather_info.Temperature_[0], _ = strconv.Atoi(temp[1])
					brief_weather_info.Temperature_[1], _ = strconv.Atoi(temp[0])
				} else {
					brief_weather_info.Temperature_[0], _ = strconv.Atoi(temp[0])
					brief_weather_info.Temperature_[1], _ = strconv.Atoi(temp[0])
				}
			case 11: //windy-1
				brief_weather_info.Wind_.From_ = wordfind_re.FindString(string(value))
			case 12: //windy-2
				brief_weather_info.Wind_.To_ = wordfind_re.FindString(string(value))
			case 13, 14: //windy-level
				lven := len(value)
				if lven < 8 {
					continue
				}
				brief_weather_info.Wind_.Level_ = string(value[3:(lven - 4)])
				break
			}
		}

		SevenDaysWeatherInfo.Weather_[localWeatherIndex] = brief_weather_info
		localWeatherIndex++
		if localWeatherIndex >= 7 {
			break
		}
	}
	/*parse live index*/
	live_index_start := live_index_re.FindStringIndex(string(body))
	live_index_end := end_re.FindStringIndex(string(body[live_index_start[0]:]))
	live_index_str := body[live_index_start[1] : live_index_start[0]+live_index_end[0]]
	liveIndexInfos := strings.Split(string(live_index_str), DAY_INFO_SPLIT_SEP)
	//airOnStr := airOn_re.FindString(string(body[live_index_end[1]:]))
	// fmt.Printf("AQI: ")
	// if "" != airOnStr {
	// 	fmt.Printf("%s\n", airOnStr)
	// } else {
	// 	fmt.Printf("无\n")
	// }
	var localLiveIndex int = 0
	for _, liveInfo := range liveIndexInfos {
		lines := strings.SplitAfter(liveInfo, "\n")
		linesLen := len(lines) - 1
		if linesLen < 4 {
			continue
		}
		star_count := len(sport_star_re.FindAllIndex([]byte(liveInfo), -1))
		lines = lines[:linesLen]
		linesLen -= 1
		var info = &LiveIndexInfo{}
		field_index := 0
		for i := 0; i < linesLen; i++ {
			strLive := lines[linesLen-i]
			strLive = strings.TrimSuffix(strLive, "\n")
			line_len := len(strLive)
			ephpmeral := wordfind_re.FindString(strLive)
			if "" == ephpmeral && i != 0 {
				for j := 0; j < star_count; j++ {
					if 0 != j {
						info.Level_ += " "
					}
					info.Level_ += "☆"
				}
				break
			} else if "" == ephpmeral && 0 == i {
				continue
			}
			switch field_index {
			case 0:
				info.Tips = strLive[3 : line_len-4]
			case 1:
				info.Name_ = strLive[4 : line_len-5]
				if 183 == info.Name_[7] {
					info.Name_ = info.Name_[8:]
				}
			case 2:
				info.Level_ = ephpmeral
			}
			field_index++
		}
		SevenDaysWeatherInfo.LiveIndex_[localLiveIndex] = info
		localLiveIndex++
		if localLiveIndex >= 6 {
			break
		}
	}
	timeNow := time.Now()
	if isFirst {
		hour, _ := strconv.Atoi(uptime[0])
		min, _ := strconv.Atoi(uptime[1])
		SevenDaysWeatherInfo.getime_ = time.Date(timeNow.Year(), timeNow.Month(), timeNow.Day(), hour, min, 0, 0, timeNow.Location())
	} else {
		SevenDaysWeatherInfo.getime_ = timeNow
	}
	c.addWeatherInfoToCache(cityinfo.Code_, SevenDaysWeatherInfo)
	return SevenDaysWeatherInfo, err
}

/*
func (c *weather) getTopList() (list string) {
	c.entryCacheMu.RLock()
	value, err := c.entryCache.Get("topentry")
	if err {
		c.entryCacheMu.RUnlock()
		return value.(string)s
	}
	log.Println("GetTopList from cache failed")
	c.weatherMu.RLock()
	defer c.weatherMu.RUnlock()

	//traversal

	return value.(string)
}
*/

func (c *Weather) TraversalRegionTree() {
	//implTraversal(c.treeRegion)
	for _, province := range c.treeRegion.Regions {
		fmt.Printf("Name:%s\tCode:%s\tUrl:%s\n", province.Name_, province.Code_, province.Url_)
		for _, citys := range province.Regions {
			fmt.Printf("    %s\n", citys.FullName_)
			for _, county := range citys.Regions {
				fmt.Printf("      └──\tName:%s\tCode:%s\tUrl:%s\n", county.FullName_, county.Code_, county.Url_)
			}
		}
	}
	// buf, _ := json.Marshal(c.treeRegion.Regions)
	// fmt.Printf("%s\n", buf)
}

func implTraversal(region *TreeRegionInfo) {
	if nil == region {
		return
	}
	fmt.Printf("Name:%s FullName:%s Url:%s Code:%s\n", region.Name_, region.FullName_, region.Url_, region.Code_)
	for _, info := range region.Regions {
		implTraversal(info)
	}
}

func (c *Weather) Stats() CacheStats {
	c.weatherMu.RLock()
	defer c.weatherMu.RUnlock()
	return CacheStats{
		Bytes:     c.nbytes,
		Items:     c.itemsLocked(),
		Gets:      c.nget,
		Hits:      c.nhit,
		Evictions: c.nevict,
	}
}

func (c *Weather) addWeatherInfoToCache(key string, value *WeatherInfo) {
	c.weatherMu.Lock()
	defer c.weatherMu.Unlock()
	if c.weatherlru.OnEvicted == nil {
		c.weatherlru.OnEvicted = func(key lrucache.Key, value interface{}) {
			val := value.(*WeatherInfo)
			c.nbytes -= int64(len(key.(string))) + int64(unsafe.Sizeof(val))
			c.nevict++
		}
	}
	c.weatherlru.Add(key, value)
	c.nbytes += int64(len(key)) + int64(unsafe.Sizeof(value))
}

func (c *Weather) getWeatherInfoForCache(key string) (value *WeatherInfo, ok bool) {
	c.weatherMu.RLock()
	defer c.weatherMu.RUnlock()
	c.nget++
	if c.weatherlru == nil {
		return
	}
	vi, ok := c.weatherlru.Get(key)
	if !ok {
		return
	}
	c.nhit++
	return vi.(*WeatherInfo), true
}

func (c *Weather) RemoveOldest() {
	c.weatherMu.Lock()
	defer c.weatherMu.Unlock()
	if c.weatherlru != nil {
		c.weatherlru.RemoveOldest()
	}
}

func (c *Weather) items() int64 {
	c.weatherMu.RLock()
	defer c.weatherMu.RUnlock()
	return c.itemsLocked()
}

func (c *Weather) itemsLocked() int64 {
	if c.weatherlru == nil {
		return 0
	}
	return int64(c.weatherlru.Len())
}

func (c *Weather) saveRegionData(path string) error {
	err := os.Remove(path)
	if err != nil {
		log.Println(err)
	}

	saveTo, err := os.Create(path)
	if err != nil {
		log.Println("Cannot create", path, err)
		return err
	}
	defer saveTo.Close()

	encoder := gob.NewEncoder(saveTo)
	err = encoder.Encode(c.treeRegion)
	if err != nil {
		log.Println("Cannot save to", path, err)
		return err
	}
	return nil
}

func (c *Weather) loadRegionData(path string) error {
	loadFrom, err := os.Open(path)
	defer loadFrom.Close()
	if err != nil {
		log.Println("Load region data from ", path, "failed", err)
		return err
	}

	decoder := gob.NewDecoder(loadFrom)
	if nil == c.treeRegion {
		c.treeRegion = &TreeRegionInfo{Regions: make(map[string]*TreeRegionInfo)}
	}
	return decoder.Decode(c.treeRegion)
}
