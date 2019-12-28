package fdfs_client

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"io"
)

type ExposeUploadConn struct {
	conn pConn
	task *storageUploadTask

	totalSize 	int64
	curSize		int64

	completeFlag	bool
	preparationFlag	bool
}

type ExposeDownloadConn struct {
	conn 	pConn
	task 	*storageDownloadTask

	preparationFlag 	bool
	completeFlag		bool

	curSize		int64
}

func (this *ExposeUploadConn) PreparationTransmission() error{
	this.task.cmd = STORAGE_PROTO_CMD_UPLOAD_FILE
	this.task.pkgLen = this.task.fileInfo.fileSize + 15

	if err := this.task.SendHeader(this.conn); err != nil {
		return err
	}

	buffer := new(bytes.Buffer)
	buffer.WriteByte(byte(this.task.storagePathIndex))
	if err := binary.Write(buffer, binary.BigEndian, this.task.fileInfo.fileSize); err != nil {
		return err
	}

	byteFileExtName := []byte(this.task.fileInfo.fileExtName)
	var bufferFileExtName [6]byte
	for i := 0; i < len(byteFileExtName); i++ {
		bufferFileExtName[i] = byteFileExtName[i]
	}
	buffer.Write(bufferFileExtName[:])

	if _, err := this.conn.Write(buffer.Bytes()); err != nil {
		return err
	}

	this.preparationFlag = true
	return nil
}

func (this *ExposeUploadConn) Write(b []byte) (n int, err error){
	if !this.preparationFlag{
		return 0, errors.New("not ready yet for transmission")
	}
	n, err = this.conn.Write(b)
	this.curSize += int64(n)
	return n, err

}

func (this *ExposeUploadConn) WriteComplete() (fileId string, err error) {
	if this.completeFlag{
		return this.task.fileId, nil
	}

	if this.curSize < this.totalSize{
		return "", errors.New(fmt.Sprintf("transmission not complete (%d < %d)", this.curSize, this.totalSize))
	}

	err = this.task.RecvRes(this.conn)
	if err != nil{
		return "", err
	}
	this.completeFlag = true
	return this.task.fileId, nil
}

func (this *ExposeUploadConn) Close() error {
	return this.conn.Close()
}


func (this *ExposeDownloadConn) Read(b []byte) (n int, err error){
	if !this.preparationFlag{
		return 0, errors.New("not ready yet for transmission")
	}
	if this.completeFlag{
		return 0, io.EOF
	}

	n, err = this.conn.Read(b)
	this.curSize += int64(n)
	if this.curSize >= this.task.pkgLen{
		this.completeFlag = true
	}
	return n, err
}

func (this *ExposeDownloadConn) PreparationTransmission() error{
	err := this.task.SendReq(this.conn)
	if err != nil{
		return err
	}

	if err := this.task.RecvHeader(this.conn); err != nil {
		return fmt.Errorf("StorageDownloadTask RecvRes %v", err)
	}

	this.preparationFlag = true
	return nil
}

func (this *ExposeDownloadConn) Close() error{
	return this.conn.Close()
}