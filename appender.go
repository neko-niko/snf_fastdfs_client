package fdfs_client

import (
	"io"
)

//var appenderMap  map[string]Appender

//func gcErrorProcess(err error){
//	fmt.Println(err)
//}
//
//func appenderClear(timeoutChan <-chan string) {
//	for {
//	select {
//	case timeoutId := <-timeoutChan:
//		if timeoutId == "" {
//			return
//		}
//		timeoutAppender := appenderMap[timeoutId]
//		err := timeoutAppender.Clear()
//		if err != nil {
//			gcErrorProcess(err)
//		}
//		delete(appenderMap, timeoutId)
//
//	}
//	}
//}

type Appender interface {
	AppendByStream(s io.Reader) error
	Clear() error
}

type appender struct {
	//fileName	string
	//groupName 	string
	fileId string
	storageAddr	string
	curNum	int
}
type storageInitTask struct {
	header
	fileinfo fileInfo
	appender *appender
}

type storageAppendTask struct {
	header
	fileinfo fileInfo
	appender *appender
}