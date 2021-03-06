package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

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

var estelle *Estelle

func main() {
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}
	estelle, err = New(opts.CacheDir)
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/status", handleStatus).
		Methods("GET")
	router.HandleFunc("/path", handlePath).
		Methods("GET")
	router.HandleFunc("/content", handleContent).
		Methods("GET")
	router.HandleFunc("/queue", handleQueue).
		Methods("POST")

	n := negroni.New(negroni.NewRecovery(), negroni.NewLogger())
	n.UseHandler(router)
	n.Run(fmt.Sprintf(":%d", opts.Port))
}

func handleStatus(res http.ResponseWriter, req *http.Request) {
	ti, err := thumbInfoFromReq(req)
	switch err.(type) {
	case nil:
		// OK
	case InvalidIdError:
		res.WriteHeader(404)
		res.Write([]byte(err.Error()))
		return
	default:
		if os.IsNotExist(err) {
			res.WriteHeader(404)
			res.Write([]byte("Not found"))
			return
		}
		panic(err)
	}
	if estelle.Exists(ti) {
		res.Header().Add("ETag", ti.ETag())
		res.WriteHeader(200)
		return
	}
	if estelle.IsInQueue(ti) {
		res.Header().Add("ETag", ti.ETag())
		res.WriteHeader(202)
		return
	}
	res.WriteHeader(404)
	res.Write([]byte("Not found"))
}

func handlePath(res http.ResponseWriter, req *http.Request) {
	path, ti, err := findOrMakeThumbFromReq(req)
	switch err.(type) {
	case nil:
		// OK
	case InvalidIdError, NoSourceError:
		res.WriteHeader(404)
		res.Write([]byte("Not found"))
		return
	default:
		if os.IsNotExist(err) {
			res.WriteHeader(404)
			res.Write([]byte("Not found"))
			return
		}
		panic(err)
	}
	res.Header().Add("Content-Type", "text/plain")
	res.Header().Add("ETag", ti.ETag())
	res.WriteHeader(200)
	fmt.Fprint(res, path)
}

func handleContent(res http.ResponseWriter, req *http.Request) {
	path, ti, err := findOrMakeThumbFromReq(req)
	switch err.(type) {
	case nil:
		// Ok
	case InvalidIdError, NoSourceError:
		res.WriteHeader(404)
		res.Write([]byte("Not found"))
		return
	default:
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
	res.Header().Add("Content-Type", ti.Format().MimeType())
	res.Header().Add("ETag", ti.ETag())
	res.WriteHeader(200)
	io.Copy(res, file)
}

func handleQueue(res http.ResponseWriter, req *http.Request) {
	ti, err := thumbInfoFromReq(req)
	if err != nil {
		if os.IsNotExist(err) {
			res.WriteHeader(404)
			res.Write([]byte("Not found"))
			return
		}
		panic(err)
	}
	res.Header().Add("ETag", ti.ETag())
	if estelle.Exists(ti) {
		res.WriteHeader(200)
		return
	}
	if !estelle.IsInQueue(ti) {
		estelle.Enqueue(5, ti)
	}
	res.WriteHeader(202) // Accepted
}

func findOrMakeThumbFromReq(req *http.Request) (string, *ThumbInfo, error) {
	ti, err := thumbInfoFromReq(req)
	if err != nil {
		return "", nil, err
	}
	path, err := estelle.Get(2, ti)
	if err != nil {
		return "", nil, err
	}
	return path, ti, nil
}

func thumbInfoFromReq(req *http.Request) (*ThumbInfo, error) {
	if len(req.URL.Query()["id"]) > 0 {
		return NewThumbInfoFromId(req.URL.Query()["id"][0])
	}
	source := ""
	if len(req.URL.Query()["source"]) > 0 {
		source = req.URL.Query()["source"][0]
	}
	if source != "" && source[0] != '/' {
		source = "/" + source
	}
	size := parseQuerySize(req.URL.Query()["size"])
	mode := parseQueryMode(req.URL.Query()["mode"])
	format := parseQueryFormat(req.URL.Query()["format"])
	return NewThumbInfoFromFile(source, size, mode, format)
}

func parseQuerySize(query []string) Size {
	if len(query) > 0 {
		size, err := SizeFromString(query[0])
		if err == nil {
			return size
		}
	}
	return SizeFromUint(85, 85)
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

func parseQueryFormat(query []string) Format {
	if len(query) > 0 {
		if format, err := FormatFromString(query[0]); err != nil {
			return format
		}
	}
	return FMT_JPG
}
