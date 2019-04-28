package main

import (
	"WeatherInfos/weather"
	"encoding/json"
	"flag"
	"fmt"
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
	LOG_FILE   = "./logs/weather.log"
	FIELD_NAME = "city"
	STR_SEP    = ","
)

var (
	iservices = flag.Bool("s", false, "To running as a services")
	port      = flag.Int("port", 3244, "The TCP port that the server listens on")
	address   = flag.String("address", "", "The net address that the server listens")
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

	getWeatherHandle()

	router := http.NewServeMux()
	router.HandleFunc("/", safe_http_handle(safe_statement))
	router.HandleFunc("/weather", safe_http_handle(ShowWeather))
	router.HandleFunc("/citylist", safe_http_handle(ShowCityList))
	router.HandleFunc("/weather/status", safe_http_handle(ShowStatus))

	fmt.Printf("Service listen on %s:%d\n", *address, *port)
	log.Printf("Service listen on %s:%d\n", *address, *port)

	go listenSignal()
	// if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), router); err != nil {
	// 	log.Println(err)
	// }
	if err := http.ListenAndServeTLS(fmt.Sprintf("%s:%d", *address, *port), "server.crt", "server.key", router); err != nil {
		log.Println(err)
	}
}

func safe_statement(w http.ResponseWriter, r *http.Request) {
	t := time.Now().Format("2006-01-02 15:04:05Z07:00")
	fmt.Fprintf(w, "<h1 align=\"center\">This is a laboratory test environment.</h1>")
	fmt.Fprintf(w, "<h3 align=\"center\">current time is</h1>")
	fmt.Fprintf(w, "<h2 align=\"center\">%s</h2>\n", t)
}

func getWeatherHandle() (weatherhandle *weather.Weather) {
	once.Do(func() {
		handle = weather.New(weather.DEFAULT_LIMIT_SIZE)
		if err := handle.InitRegionTree(); err != nil {
			log.Println(err)
		}
	})
	return handle
}

func ShowStatus(w http.ResponseWriter, r *http.Request) {
	weatherHandle := getWeatherHandle()
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

	weatherHandle := getWeatherHandle()
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
		http.Error(w, http.ErrBodyNotAllowed.Error(), http.StatusMethodNotAllowed)
		return
	}
	r.ParseForm()
	prov, has := r.Form[FIELD_NAME]
	if !has {
		http.Error(w, "parameter error", http.StatusBadRequest)
		return
	}
	strCity := prov[0]
	params := strings.Split(strCity, STR_SEP)
	paramslen := len(params)
	weatherHandle := getWeatherHandle()
	var err error
	var Resp *weather.WeatherInfo

	if nil == weatherHandle {
		log.Printf("weatherHandle is nil, please check")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	switch paramslen {
	case 3:
		Resp, err = weatherHandle.ShowCityWeather(params[0], params[1], params[2])
	case 2:
		Resp, err = weatherHandle.ShowCityWeather(params[0], params[1], params[1])
	case 1:
		Resp, err = weatherHandle.ShowCityWeather(params[0], params[0], params[0])
	default:
		http.Error(w, "parameter error", http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	if nil != err {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if nil == Resp {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	jsonstr, err := json.Marshal(Resp)
	if nil != err {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	w.Write(jsonstr)
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
	fmt.Println("     -s\tSet process running as a services, using [false] by default")
	fmt.Println("     -address\tSet the listener address, using [0.0.0.0] by default")
	fmt.Println("     -port\tSet the listener port, using port [3244] by default")
	fmt.Println("     -help\tdisplay help info and exit")
}

func runAsServices() {
	if *iservices {
		cmd := exec.Command(os.Args[0], flag.Args()...)
		cmd.Start()
		fmt.Printf("%s [PID] %d running...\n", filepath.Base(os.Args[0]), cmd.Process.Pid)
		log.Printf("%s [PID] %d running...\n", filepath.Base(os.Args[0]), cmd.Process.Pid)
		*iservices = false
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
