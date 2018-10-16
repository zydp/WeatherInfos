package main

import (
	"WeatherInfos/weather"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// /*
// #include <unistd.h>
// */
// import "C"

const (
	LOG_FILE   = "./logs/weather.log"
	FIELD_NAME = "city"
	STR_SEP    = "_"
)

var (
	port    = flag.Int("port", 3244, "The TCP port that the server listens on")
	address = flag.String("address", "", "The net address that the server listens")
	handle  *weather.Weather
	once    sync.Once
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
	//C.daemon(1,1)  /*when the background runs, open this line, or you will run it as 'nohup ./WeatherInfos 2>&1 &' */
	flag.Parse()
	if flag.NFlag() <= 0 {
		fmt.Printf("using default setting, listen on %s:%d\n", *address, *port)
		log.Printf("using default setting, listen on %s:%d\n", *address, *port)
	}
	getWeatherHandle()
	router := http.NewServeMux()
	router.HandleFunc("/", safe_http_handle(safe_statement))
	router.HandleFunc("/weather", safe_http_handle(ShowWeather))
	router.HandleFunc("/weather/status", safe_http_handle(ShowStatus))
	if err := http.ListenAndServe(fmt.Sprintf("%s:%d", *address, *port), router); err != nil {
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
	fmt.Printf("Usage: %s [OPTION]...\n", os.Args[0])
	fmt.Println("     -address\tSet the listener address, use 0.0.0.0 by default")
	fmt.Println("     -port\tSet the listener port, use port 3244 by default")
	fmt.Println("     -help\tdisplay help info and exit")
}


/*  //using C.daemon
func daemon(nochdir, noclose int) int {
    var ret, ret2 uintptr
    var err syscall.Errno
 
    darwin := runtime.GOOS == "darwin"
 
    // already a daemon
    if syscall.Getppid() == 1 {
        return 0
    }
 
    // fork off the parent process
    ret, ret2, err = syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
    if err != 0 {
        return -1
    }
 
    // failure
    if ret2 < 0 {
        os.Exit(-1)
    }
 
    // handle exception for darwin
    if darwin && ret2 == 1 {
        ret = 0
    }
 
    // if we got a good PID, then we call exit the parent process.
    if ret > 0 {
        os.Exit(0)
    }
 
    // Change the file mode mask 
    _ = syscall.Umask(0)
 
    // create a new SID for the child process
    s_ret, s_errno := syscall.Setsid()
    if s_errno != nil {
        log.Printf("Error: syscall.Setsid errno: %d", s_errno)
    }
    if s_ret < 0 {
        return -1
    }
 
    if nochdir == 0 {
        os.Chdir("/")
    }
 
    if noclose == 0 {
        f, e := os.OpenFile("/dev/null", os.O_RDWR, 0)
        if e == nil {
            fd := f.Fd()
            syscall.Dup2(int(fd), int(os.Stdin.Fd()))
            syscall.Dup2(int(fd), int(os.Stdout.Fd()))
            syscall.Dup2(int(fd), int(os.Stderr.Fd()))
        }
    }
    return 0
}
*/
