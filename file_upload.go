package main

import (
	"log"
	"net/http"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	//"./fdfs_client"
)
const(
	maxUploadSize=1024*1024*500
)
func renderError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(message))
}

func uploadFileHandler()(http.HandlerFunc){
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

		if err := r.ParseMultipartForm(maxUploadSize); err != nil{
			log.Println(err)
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		}
		log.Println(r.PostForm["token"])

		fileSize := r.Header["Filesize"][0]
		filesize, _ := strconv.Atoi(fileSize)
		fmt.Println("filesize: ", filesize)

		fileType := r.PostFormValue("token")
		log.Println(fileType)
		file, _, err := r.FormFile("uploadFile")
		if err != nil{
			log.Println(err)
			renderError(w, "INVALID_FILE", http.StatusBadRequest)
			return
		}

		defer file.Close()



		s := make([]byte, 4096)
		file_name := "testfile"
		work_dir, _ := filepath.Abs(".")
		local_file, err := os.OpenFile(
			filepath.Join(work_dir, file_name+"."+fileType),
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			0666,
		)
		if err != nil{
			renderError(w, "FILE_OPEN_ERROR", http.StatusServiceUnavailable)
			return
		}

		defer local_file.Close()

		for {
			switch nr, err := file.Read(s); true {
			case nr < 0:
				fmt.Fprintf(os.Stderr, "cat: error reading: %s\n", err.Error())
				return
			case nr == 0:
				return
			default:
				_, err := local_file.Write(s[:nr])
				if err != nil{
					renderError(w, "CANT_WRITE_FILE", http.StatusInternalServerError)
					return
				}
			}
		}
		w.Write([]byte("SUCCESS"))
	})
}


func main(){
	http.HandleFunc("/upload", uploadFileHandler())
	fs := http.FileServer(http.Dir("."))
	http.Handle("/files/", http.StripPrefix("/files", fs))
	log.Print("Server started on localhost:8080, use /upload for uploading files and /files/{fileName} for downloading files.")

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
