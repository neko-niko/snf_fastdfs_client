package fdfs_client

import (
	"io"
	"log"
	"mime"
	"net/http"
	"os"
)

var client *Client

func init(){
	var err error
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	client, err = NewClientWithConfig("./configs/fdfs.conf")
	if err != nil{
		log.Println(err)
		return
	}
}

func uploadInterceptorTest(){
	file, err := os.Open("./test.jpg")
	if err != nil{
		log.Println(err)
		return
	}

	fileStat, err := file.Stat()
	if err != nil{
		log.Println(err)
		return
	}
	total := fileStat.Size()

	conn, err := client.InterceptUpload("", total)
	defer conn.Close()
	if err != nil{
		log.Println(err)
	}

	err = conn.PreparationTransmission()
	if err != nil{
		log.Println(err)
	}

	typeBytes := make([]byte, 512)
	n1, err := file.Read(typeBytes)
	if err != nil{
		log.Println(err)
	}
	mimeType := http.DetectContentType(typeBytes[:n1])
	log.Println(mimeType)
	extName, _ := mime.ExtensionsByType(mimeType)
	log.Println(extName)

	_, err = conn.Write(typeBytes)
	if err != nil{
		log.Println(err)
	}

	n2, err := io.Copy(conn, file)
	if err != nil{
		log.Println(err)
	}

	log.Println(total, n1, n2)

	fileId, err := conn.WriteComplete()
	if err != nil{
		log.Println(err)
	}
	log.Println(fileId)
}

func testDownloadIntercept(){
	conn, err := client.InterceptDownload("group1/M00/00/00/rBEAAV4HGCqAQjVJAAPOwEFyXZ01830344", 0, 0)
	defer conn.Close()
	if err != nil{
		log.Println(err)
	}

	err = conn.PreparationTransmission()
	if err != nil{
		log.Println(err)
	}

	typebytes := make([]byte, 512)
	n1, err := conn.Read(typebytes)
	if err != nil{
		log.Println(err)
	}
	mimetype := http.DetectContentType(typebytes)
	extnames, err := mime.ExtensionsByType(mimetype)
	extname := extnames[0]
	log.Println(mimetype, extname)

	file, err := os.Create("test"+extname)
	defer file.Close()
	if err != nil{
		log.Println(err)
	}
	_, err = file.Write(typebytes)
	if err != nil{
		log.Println(err)
	}
	n2, err := io.Copy(file, conn)
	if err != nil{
		log.Println(err)
	}

	log.Println(n1, n2)

}

func Test(){


	testDownloadIntercept()



	//stream := &streamInfo{
	//	stream:     s,
	//	streamSize: total,
	//}
	//
	//fileInfo, err := newFileInfo("", nil, stream, "")
	//if err != nil{
	//	log.Println(err)
	//}
	//
	//
	//storageinfo, err := client.queryStorageInfoWithTracker(TRACKER_PROTO_CMD_SERVICE_QUERY_STORE_WITHOUT_GROUP_ONE, "", "")
	//if err != nil{
	//	log.Println(err)
	//	return
	//}
	//
	//task := &storageUploadTask{}
	//task.fileInfo = fileInfo
	//task.storagePathIndex = storageinfo.storagePathIndex
	//
	//storageConn, err := client.getStorageConn(storageinfo)
	//if err != nil{
	//	log.Println(err)
	//	return
	//}
	//
	//err = task.SendReq(storageConn)
	//if err != nil{
	//	log.Println("send err")
	//	log.Println(err)
	//}
	//
	//err = task.RecvRes(storageConn)
	//if err != nil{
	//	log.Println("recv err")
	//	log.Println(err)
	//}
	//
	//log.Println(task.fileId)
	//fileInfo.Close()
	//
	//storageConn.Close()

}
