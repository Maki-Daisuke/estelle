package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	router.HandleFunc("/file{path:(/[^?]+)?}", handleFile).
		Methods("GET")
	router.HandleFunc("/thumb{path:(/[^?]+)?}", handleThumb).
		Methods("GET")

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(router)
	n.Run(fmt.Sprintf(":%d", opts.Port))
}

func handleFile(res http.ResponseWriter, req *http.Request) {
	path, err := findOrMakeThumbnail(req)
	if err != nil {
		if os.IsNotExist(err) {
			res.WriteHeader(404)
			res.Write([]byte("Not found"))
			return
		}
		panic(err)
	}
	res.Header().Add("Content-Type", "text/plain")
	res.WriteHeader(200)
	fmt.Fprint(res, path)
}

func handleThumb(res http.ResponseWriter, req *http.Request) {
	path, err := findOrMakeThumbnail(req)
	if err != nil {
		if os.IsNotExist(err) {
			res.WriteHeader(404)
			res.Write([]byte("Not found"))
			return
		}
		panic(err)
	}
	file, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	mimeType := "image/jpeg"
	switch filepath.Ext(path) {
	case ".png":
		mimeType = "image/png"
	case ".webp":
		mimeType = "image/webp"
	}
	res.Header().Add("Content-Type", mimeType)
	res.WriteHeader(200)
	io.Copy(res, file)
}

func findOrMakeThumbnail(req *http.Request) (string, error) {
	vars := mux.Vars(req)
	path := vars["path"]
	if path == "" {
		if len(req.URL.Query()["path"]) == 0 {
			return "", fmt.Errorf("")
		}
		path = req.URL.Query()["path"][0]
		if len(path) == 0 || path[0] != '/' {
			path = "/" + path
		}
	}
	width, height := parseQuerySize(req.URL.Query()["size"])
	mode := parseQueryMode(req.URL.Query()["mode"])
	format := parseQueryFormat(req.URL.Query()["format"])
	ti, err := NewThumbInfoFromFile(path, width, height, mode, format)
	if err != nil {
		return "", err
	}
	path, err = cacheDir.Get(ti)
	if err != nil {
		return "", err
	}
	return path, nil
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
