package fdfs_client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"strings"
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
	AppendByStream(s io.Reader, total int64) error
	Clear() error
	Complete() (string, error)
}

type appender struct {
	//fileName	string
	//groupName 	string
	fileId string
	storageAddr	string
	pathIndex 	int8
	fileExtName	string
	fatherC	*Client
}

func (this *appender) Clear() error {
	return nil
}

func (this *appender) Complete() (string, error){
	return this.fileId, nil
}

func (this *appender) AppendByStream(s io.Reader, total int64) error {
	stream := &streamInfo{stream:s, streamSize:total}


	var err error
	if this.storageAddr == "" {
		err = this.appenderInit(stream)
	} else{
		err = this.append(stream)
	}

	if err != nil{
		return err
	}

	return nil
}

func (this *appender) appenderInit(s *streamInfo) error{
	fileInfo, err := newFileInfo("", nil, s, this.fileExtName)
	if err != nil{
		return err
	}
	defer fileInfo.Close()

	storageInfo, err := this.fatherC.queryStorageInfoWithTracker(TRACKER_PROTO_CMD_SERVICE_QUERY_STORE_WITHOUT_GROUP_ONE, "", "" )
	if err != nil{
		return err
	}
	this.storageAddr = storageInfo.addr
	this.pathIndex = storageInfo.storagePathIndex

	task := &storageAppenderInitTask{}
	task.appender = this
	task.fileInfo = fileInfo

	err = this.doStorage(task)
	if err != nil{
		return err
	}

	this.fileId = task.fileId
	return nil
}

func (this *appender) append(s *streamInfo) error{
	fileInfo, err := newFileInfo("", nil, s, this.fileExtName)
	if err != nil{
		return err
	}
	defer fileInfo.Close()
	task := &storageAppendTask{}
	task.appender = this
	task.fileInfo = fileInfo
	err = this.doStorage(task)
	if err != nil{
		return err
	}
	return nil

}

func (this *appender) doStorage(task task) error{
	tempStorageInfo := &storageInfo{
		addr:             this.storageAddr,
		storagePathIndex: this.pathIndex,
	}
	err := this.fatherC.doStorage(task, tempStorageInfo)
	if err != nil{
		return err
	}
	return nil
}




type storageAppenderInitTask struct {
	header
	fileInfo *fileInfo
	appender *appender
	fileId string
}

func (this *storageAppenderInitTask) SendReq(conn net.Conn) error{
	this.cmd = STORAGE_PROTO_CMD_UPLOAD_APPENDER_FILE
	this.pkgLen = this.fileInfo.fileSize + 15

	err := this.SendHeader(conn)
	if err != nil{
		return err
	}

	// build metadata(pathindex, filesize, extname)
	buffer := new(bytes.Buffer)
	buffer.WriteByte(byte(this.appender.pathIndex))
	err = binary.Write(buffer, binary.BigEndian, this.fileInfo.fileSize)
	if err != nil{
		return err
	}
	fileExtName := []byte(this.appender.fileExtName)
	if len(fileExtName) < 6{
		extNameExpend := make([]byte, 6-len(fileExtName))
		fileExtName = append(fileExtName, extNameExpend...)
	}
	buffer.Write(fileExtName)

	_, err = conn.Write(buffer.Bytes())
	if err != nil{
		return err
	}

	 // send data
	if this.fileInfo.file != nil{
		// do something to send file
	} else if this.fileInfo.buffer != nil{
		// do something to send buffer
	} else{
		_, err = io.Copy(conn, this.fileInfo.streaminfo.stream)
	}

	if err != nil{
		return err
	}
	return nil

}

func (this *storageAppenderInitTask) RecvRes(conn net.Conn) error{
	err := this.RecvHeader(conn)
	if err != nil{
		return err
	}

	if this.pkgLen <= 16 {
		return errors.New("recv file id pkgLen <= FDFS_GROUP_NAME_MAX_LEN")
	}
	if this.pkgLen > 100 {
		return errors.New("recv file id pkgLen > 100,can't be so long")
	}

	buf := make([]byte, this.pkgLen)
	_, err  = conn.Read(buf)
	if err != nil{
		return err
	}

	buffer := bytes.NewBuffer(buf)
	groupName, err := readCStrFromByteBuffer(buffer, 16)
	if err != nil{
		return err
	}

	remoteFileName, err := readCStrFromByteBuffer(buffer, int(this.pkgLen-16))
	if err != nil{
		return err
	}
	this.fileId = groupName + "/" + remoteFileName
	return nil

}

type storageAppendTask struct {
	header
	fileInfo *fileInfo
	appender *appender
}

func (this *storageAppendTask) SendReq(conn net.Conn) error{
	//tempStorageInfo := &storageInfo{
	//	addr:             this.appender.storageAddr,
	//	storagePathIndex: this.appender.pathIndex,
	//}
	fileName := []byte(strings.SplitN(this.appender.fileId, "/", 2)[1])
	fileNameLen := int64(len(fileName))

	this.cmd = STORAGE_PROTO_CMD_APPEND_FILE
	this.pkgLen = fileNameLen + this.fileInfo.fileSize + 16
	err := this.SendHeader(conn)
	if err != nil{
		return err
	}

	buffer := new(bytes.Buffer)
	err = binary.Write(buffer, binary.BigEndian, fileNameLen)
	if err != nil{
		return err
	}
	err = binary.Write(buffer, binary.BigEndian, this.fileInfo.fileSize)
	if err != nil{
		return err
	}

	_, err = buffer.Write(fileName)
	if err != nil{
		return err
	}

	_, err = conn.Write(buffer.Bytes())
	if err != nil{
		return err
	}

	if this.fileInfo.file != nil{
		// do something to send file
	} else if this.fileInfo.buffer != nil{
		// do something to send buffer
	} else{
		_, err = io.Copy(conn, this.fileInfo.streaminfo.stream)
	}

	if err != nil{
		return err
	}

	return nil


}

func (this *storageAppendTask) RecvRes(conn net.Conn) error{
	err := this.RecvHeader(conn)
	if err != nil{
		return err
	}
	return nil
}