package main

import (
	"WeatherInfos/weather"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mozillazg/go-pinyin"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
)

const (
	LOG_FILE        = "./logs/weather.log"
	FIELD_NAME      = "city"
	FIELD_NAME_CODE = "cityCode"
	STR_SEP         = ","
)

var (
	iServices = flag.Bool("s", false, "To running as a services")
	port      = flag.Int("port", 3244, "The TCP port that the server listens on")
	address   = flag.String("address", "", "The net address that the server listens")
	crt       = flag.String("crt", "", "Specify the server credential file")
	key       = flag.String("key", "", "Specify the server key file")
	handle    *weather.Weather
	once      sync.Once
	sigs      = make(chan os.Signal, 1)
	exit      = make(chan bool, 1)
)

func init() {
	flag.CommandLine.Usage = help
	if logout, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666); err == nil {
		log.SetOutput(logout)
		log.SetPrefix("[Info] ")
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.Println(time.Now().Format(time.RFC3339), strings.Title(runtime.GOARCH), strings.Title(runtime.GOOS))
	} else {
		fmt.Println(time.Now().Format(time.RFC3339), strings.Title(runtime.GOARCH), strings.Title(runtime.GOOS))
		fmt.Printf("Not found the [ logs ] directory, the log will be displayed on the terminal\n")
	}
}

func main() {
	if !flag.Parsed() {
		flag.Parse()
	}
	runAsServices()
	if flag.NFlag() <= 0 {
		fmt.Printf("Using default setting, listen on %s:%d\n", *address, *port)
		log.Printf("Using default setting, listen on %s:%d\n", *address, *port)
	}

	//根据设定的间隔去进行告警列表的获取
	go weather.CheckAlarmListFromWeatherCom()

	GetWeatherHandle()

	router := http.NewServeMux()
	router.HandleFunc("/", safe_http_handle(safe_statement))
	router.HandleFunc("/weather", safe_http_handle(ShowWeather))
	router.HandleFunc("/weather/forty", safe_http_handle(ShowFortyWeather))
	router.HandleFunc("/citylist", safe_http_handle(ShowCityList))
	router.HandleFunc("/weather/status", safe_http_handle(ShowStatus))

	fmt.Printf("Service listen on %s:%d\n", *address, *port)
	log.Printf("Service listen on %s:%d\n", *address, *port)

	go listenSignal()
	if "" == *crt {
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), router); err != nil {
			log.Println(err)
		}
	} else {
		if err := http.ListenAndServeTLS(fmt.Sprintf("%s:%d", *address, *port), *crt, *key, router); err != nil {
			log.Println(err)
		}
	}
}

func safe_statement(w http.ResponseWriter, r *http.Request) {
	t := time.Now().Format("2006-01-02 15:04:05Z07:00")
	fmt.Fprintf(w, "<h1 align=\"center\">This is a laboratory test environment.</h1>")
	fmt.Fprintf(w, "<h3 align=\"center\">current time is</h1>")
	fmt.Fprintf(w, "<h2 align=\"center\">%s</h2>\n", t)
}

func GetWeatherHandle() (weatherhandle *weather.Weather) {
	once.Do(func() {
		handle = weather.New(int(weather.DEFAULT_LIMIT_SIZE))
		if err := handle.InitRegionTree(); err != nil {
			log.Println(err)
		}
	})
	return handle
}

func ShowStatus(w http.ResponseWriter, r *http.Request) {
	weatherHandle := GetWeatherHandle()
	status := weatherHandle.Stats()
	w.Header().Add("Content-Type", "application/json")
	strStatus, _ := json.Marshal(status)
	w.Write(strStatus)
}

func ShowCityList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, http.ErrBodyNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	r.ParseForm()

	weatherHandle := GetWeatherHandle()
	if prov, ok := r.Form[FIELD_NAME]; ok {
		resp, _ := weatherHandle.ShowCityList(prov[0])
		w.Write(resp)
	} else {
		resp, _ := weatherHandle.ShowCityList("")
		w.Write(resp)
	}

}

func ShowWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errResp(w, http.StatusMethodNotAllowed, http.ErrBodyNotAllowed.Error())
		return
	}
	r.ParseForm()

	prov, has := r.Form[FIELD_NAME]
	//cityCode, hasCode := r.Form[FIELD_NAME_CODE]
	if !has {
		errResp(w, http.StatusBadRequest, "parameter error")
		return
	}

	strCity := prov[0]
	params := strings.Split(strCity, STR_SEP)
	var spellParams []string = make([]string, 0)

	for i := 0; i < len(params); i++ {
		var spellStrCity = ""
		for _, v := range pinyin.LazyConvert(params[i], nil) {
			spellStrCity += v
		}
		spellParams = append(spellParams, spellStrCity)

	}
	paramsLen := len(params)

	for k, v := range spellParams {
		if "" == v {
			spellParams[k] = params[k]
		}
	}

	weatherHandle := GetWeatherHandle()
	var err error
	var Resp *weather.WeatherInfo

	if nil == weatherHandle {
		log.Printf("weatherHandle is nil, please check")
		errResp(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	switch paramsLen {
	case 3:
		Resp, err = weatherHandle.ShowCityWeather(spellParams[0], spellParams[1], spellParams[2])
	case 2:
		Resp, err = weatherHandle.ShowCityWeather(spellParams[0], spellParams[1], spellParams[1])
	case 1:
		Resp, err = weatherHandle.ShowCityWeather(spellParams[0], spellParams[0], spellParams[0])
	default:
		errResp(w, http.StatusBadRequest, "parameter error")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	if nil != err {
		errResp(w, http.StatusBadRequest, err.Error())
		return
	}
	if nil == Resp {
		errResp(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	jsonStr, err := json.Marshal(Resp)
	if nil != err {
		errResp(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	w.Write(jsonStr)
}

func ShowFortyWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		errResp(w, http.StatusMethodNotAllowed, http.ErrBodyNotAllowed.Error())
		return
	}
	r.ParseForm()

	prov, has := r.Form[FIELD_NAME]
	if !has {
		errResp(w, http.StatusBadRequest, "parameter error: missing city parameter")
		return
	}

	strCity := prov[0]
	if strCity == "" {
		errResp(w, http.StatusBadRequest, "parameter error: empty city parameter")
		return
	}

	params := strings.Split(strCity, STR_SEP)
	var spellParams []string = make([]string, 0)

	for i := 0; i < len(params); i++ {
		var spellStrCity = ""
		for _, v := range pinyin.LazyConvert(params[i], nil) {
			spellStrCity += v
		}
		spellParams = append(spellParams, spellStrCity)
	}
	paramsLen := len(params)

	for k, v := range spellParams {
		if "" == v {
			spellParams[k] = params[k]
		}
	}

	weatherHandle := GetWeatherHandle()
	if nil == weatherHandle {
		log.Printf("weatherHandle is nil, please check")
		errResp(w, http.StatusInternalServerError, "Internal Server Error: weather service unavailable")
		return
	}

	var err error
	var Resp []weather.FortyDaysInfo

	switch paramsLen {
	case 3:
		Resp, err = weatherHandle.GetFortyDaysInfoWeatherCom(spellParams[0], spellParams[1], spellParams[2])
	case 2:
		Resp, err = weatherHandle.GetFortyDaysInfoWeatherCom(spellParams[0], spellParams[1], spellParams[1])
	case 1:
		Resp, err = weatherHandle.GetFortyDaysInfoWeatherCom(spellParams[0], spellParams[0], spellParams[0])
	default:
		errResp(w, http.StatusBadRequest, fmt.Sprintf("parameter error: invalid number of parameters (%d)", paramsLen))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	
	if err != nil {
		log.Printf("Error fetching weather data: %v", err)
		errResp(w, http.StatusBadRequest, fmt.Sprintf("failed to fetch weather data: %v", err))
		return
	}
	
	if Resp == nil || len(Resp) == 0 {
		log.Printf("No weather data available for city: %s", strCity)
		errResp(w, http.StatusNotFound, "no weather data available for the specified city")
		return
	}

	jsonStr, err := json.Marshal(Resp)
	if err != nil {
		log.Printf("Error marshaling response: %v", err)
		errResp(w, http.StatusInternalServerError, "Internal Server Error: failed to process weather data")
		return
	}

	w.Write(jsonStr)
}

func errResp(w http.ResponseWriter, rCode int, rMsg string) {
	var Jmap = make(map[string]interface{})
	Jmap[weather.RESP_RCODE_FIELD] = rCode
	Jmap[weather.RESP_RMSG_FIELD] = rMsg
	strResp, _ := json.Marshal(Jmap)
	w.Write(strResp)
}

func safe_http_handle(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err, ok := recover().(error)
			if ok {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}()
		fn(w, r)
	}
}

func help() {
	fmt.Printf("Provide weather access interface based on laboratory environment.\n")
	fmt.Printf("Usage: %s [OPTION]...\n", filepath.Base(os.Args[0]))
	fmt.Println("     -s\t\tSet process running as a services, using [false] by default")
	fmt.Println("     -address\tSet the listener address, using [0.0.0.0] by default")
	fmt.Println("     -port\tSet the listener port, using port [3244] by default")
	fmt.Println("     -crt\tSpecify the server credential file")
	fmt.Println("     -key\tSpecify the server key file")
	fmt.Println("     -help\tdisplay help info and exit")
}

func runAsServices() {
	if *iServices {
		cmd := exec.Command(os.Args[0], flag.Args()...)
		cmd.Start()
		fmt.Printf("%s [PID] %d running...\n", filepath.Base(os.Args[0]), cmd.Process.Pid)
		log.Printf("%s [PID] %d running...\n", filepath.Base(os.Args[0]), cmd.Process.Pid)
		*iServices = false
		os.Exit(0)
	}
}

func handleSignals(signal os.Signal) {
	log.Println("Recv a signal:", signal)
	exit <- true
	os.Exit(0)
}

func listenSignal() {
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
	for {
		sig := <-sigs
		handleSignals(sig)
	}
}
