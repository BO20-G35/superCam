package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

func main() {
	Unpackfirmware()
	time.Sleep(1 * time.Second)
	/*
		log.Println("listening on http://127.0.0.1:8082")
		http.Handle("/", handlers())
		err := http.ListenAndServe(":8082", nil)
		if err != nil {
			panic(err)
		}
	*/
}

func handlers() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", indexPage).Methods("GET")
	router.HandleFunc("/media/{mId:[0-9]+}/stream/", streamHandler).Methods("GET")
	router.HandleFunc("/media/{mId:[0-9]+}/stream/{segName:index[0-9]+.ts}", streamHandler).Methods("GET")
	router.HandleFunc("/settings/", settingsPage).Methods("GET")
	router.HandleFunc("/upload", uploadFirmware).Methods("POST")
	return router
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "index.html")
}

func settingsPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "settings.html")
}

func streamHandler(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	mId, err := strconv.Atoi(vars["mId"])
	if err != nil {
		response.WriteHeader(http.StatusNotFound)
		return
	}

	segName, ok := vars["segName"]
	if !ok {
		mediaBase := getMediaBase(mId)
		m3u8Name := "index.m3u8"
		serveHlsM3u8(response, request, mediaBase, m3u8Name)
	} else {
		mediaBase := getMediaBase(mId)
		serveHlsTs(response, request, mediaBase, segName)
	}
}

func getMediaBase(mId int) string {
	mediaRoot := "assets/media"
	return fmt.Sprintf("%s/%d", mediaRoot, mId)
}

func serveHlsM3u8(w http.ResponseWriter, r *http.Request, mediaBase, m3u8Name string) {
	mediaFile := fmt.Sprintf("%s/hls/%s", mediaBase, m3u8Name)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "application/x-mpegURL")
}

func serveHlsTs(w http.ResponseWriter, r *http.Request, mediaBase, segName string) {
	mediaFile := fmt.Sprintf("%s/hls/%s", mediaBase, segName)
	http.ServeFile(w, r, mediaFile)
	w.Header().Set("Content-Type", "video/MP2T")
}

func uploadFirmware(w http.ResponseWriter, r *http.Request) {

	_ = r.ParseMultipartForm(32 << 20)

	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, "error")
	}

	f, err := os.OpenFile("./assets/firmware.bin", os.O_WRONLY|os.O_CREATE, 0666)
	defer f.Close()

	_, _ = io.Copy(f, file)

	http.ServeFile(w, r, "index.html")
}
