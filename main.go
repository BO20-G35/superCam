package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

//structs for html files
type errorData struct {
	ErrorMsg string
}

type indexData struct {
	SettingsURL string
	StreamURL   string
}

type settingsData struct {
	UploadURL string
}

func main() {
	//run all init script because this is not done by docker
	RunMalware()

	if err := CheckDependencies(); err != nil {
		log.Println("missing dependencies")
		panic(err)
	}

	log.Println("listening on http://127.0.0.1:8082")
	http.Handle("/", handlers())
	err := http.ListenAndServe(":8082", nil)
	if err != nil {
		panic(err)
	}
}

func handlers() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", indexPage).Methods("GET")
	router.HandleFunc("/media/{mId:[0-9]+}/stream/", streamHandler).Methods("GET")
	router.HandleFunc("/media/{mId:[0-9]+}/stream/{segName:index[0-9]+.ts}", streamHandler).Methods("GET")
	router.HandleFunc("/settings/", settingsPage).Methods("GET")
	router.HandleFunc("/upload", uploadFirmware).Methods("POST", "GET")
	return router
}

func indexPage(w http.ResponseWriter, r *http.Request) {
	data := indexData{
		"http://127.0.0.1:8082/settings/",
		"http://127.0.0.1:8082/media/1/stream/",
	}
	tmpl := template.Must(template.ParseFiles("index.html"))
	tmpl.Execute(w, data)
}

func settingsPage(w http.ResponseWriter, r *http.Request) {
	data := settingsData{
		"http://127.0.0.1:8082/upload",
	}
	tmpl := template.Must(template.ParseFiles("settings.html"))
	tmpl.Execute(w, data)
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

//this function might be a bit big
//but in my defence most of the code is error handling spaghetti code
func uploadFirmware(w http.ResponseWriter, r *http.Request) {
	fmt.Println("uploadFirmware called")
	data := errorData{""}
	tmpl := template.Must(template.ParseFiles("error.html"))

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data.ErrorMsg = "upload failed: " + err.Error()
		tmpl.Execute(w, data)
		return
	}

	file, _, err := r.FormFile("uploadfile")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data.ErrorMsg = "upload failed: " + err.Error()
		tmpl.Execute(w, data)
		return
	}

	f, err := os.OpenFile(FirmwarePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data.ErrorMsg = "upload failed: " + err.Error()
		tmpl.Execute(w, data)
		return
	}

	defer f.Close()
	_, err = io.Copy(f, file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data.ErrorMsg = "upload failed: " + err.Error()
		tmpl.Execute(w, data)
		return
	}

	//start unpacking

	err = Unpackfirmware()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data.ErrorMsg = "upload failed: " + err.Error()
		tmpl.Execute(w, data)
		return
	}

	err = Unsquash()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data.ErrorMsg = "upload failed: " + err.Error()
		tmpl.Execute(w, data)
		return
	}

	err = CopyInitScripts()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		data.ErrorMsg = "upload failed: " + err.Error()
		tmpl.Execute(w, data)
		return
	}

	CleanUp()
	http.ServeFile(w, r, "index.html")

}
