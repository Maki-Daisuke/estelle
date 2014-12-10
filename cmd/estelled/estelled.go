package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"

	. "github.com/Maki-Daisuke/estelle"

	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Port     uint   `short:"p" long:"port" default:"1186" description:"Port number to listen"`
	CacheDir string `short:"d" long:"cache-dir" default:"./estelled-cache" description:"Directory to store cache data"`
	Expires  uint   `short:"E" long:"expires" default:"0" description:"How many minutes to keep thumbnail caches from its last access time (zero means no expiration)"`
	Limit    uint   `short:"L" long:"limit" default:"0" description:"How much disk space can be consumed to keep thumbnail cache"`
}

var cacheDir *CacheDir

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
	cacheDir, err = NewCacheDir(opts.CacheDir)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/file/{path:[^?]+}", handleFile).
		Methods("GET")
	router.HandleFunc("/thumb/{path}", handleThumb).
		Methods("GET")

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(router)
	n.Run(fmt.Sprintf(":%d", opts.Port))
}

func handleFile(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	width, height := parseQuerySize(req.URL.Query()["size"])
	mode := parseQueryMode(req.URL.Query()["mode"])
	format := parseQueryFormat(req.URL.Query()["format"])
	ti, err := NewThumbInfoFromFile("/"+vars["path"], width, height, mode, format)
	if err != nil {
		if os.IsNotExist(err) {
			res.WriteHeader(404)
			res.Write([]byte("Not found"))
			return
		}
		panic(err)
	}
	path, err := cacheDir.Get(ti)
	if err != nil {
		panic(err)
	}
	res.Header().Add("Content-Type", "text/plain")
	res.WriteHeader(200)
	fmt.Fprint(res, path)
}

func handleThumb(res http.ResponseWriter, req *http.Request) {
	panic("unimplemeted yet")
}

var reInt, _ = regexp.Compile("[0-9]+")
var reSize, _ = regexp.Compile("([0-9]+)x([0-9]+)")

func parseQuerySize(query []string) (width, height uint) {
	if len(query) > 0 {
		size := query[0]
		s := reInt.FindString(size)
		if s != "" {
			n, _ := strconv.ParseUint(s, 10, 32)
			return uint(n), uint(n)
		}
		m := reSize.FindStringSubmatch(size)
		if m != nil {
			w, _ := strconv.ParseUint(m[1], 10, 32)
			h, _ := strconv.ParseUint(m[2], 10, 32)
			return uint(w), uint(h)
		}
	}
	return 85, 85
}

func parseQueryMode(query []string) Mode {
	mode := ""
	if len(query) > 0 {
		mode = query[0]
	}
	switch mode {
	default:
		return ModeFill
	case "fit":
		return ModeFit
	case "shrink":
		return ModeShrink
	}
}

func parseQueryFormat(query []string) string {
	format := ""
	if len(query) > 0 {
		format = query[0]
	}
	switch format {
	default:
		return "jpg"
	case "png":
		return "png"
	case "webp":
		return "webp"
	}
}
