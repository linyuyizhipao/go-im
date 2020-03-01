package cusBufio

import (
	"bufio"
	"io"
	"os"
)

//表演的是将一个文件的内容通过指定大小的slice读取，并通过bufio这个介质写入到另外一个文件
//按切片大小去读内容到切片，将切片内容读到缓冲，满了自己会映射到对应文件，最后将缓冲未满的部分也同步到文件，这就是这个函数做的一些事情
func bufferReadWrite(path string,writePath string) error {
	file,err :=os.Open(path) //得到一个file对象
	if err != nil {
		return err
	}
	defer file.Close()

	writeFile,err := os.OpenFile(writePath,os.O_RDWR,0600) //获取一个自定义权限的file对象
	if err != nil {
		return err
	}
	defer writeFile.Close()

	//得到一个reader对象
	reader :=bufio.NewReader(file)
	bwrite := bufio.NewWriter(writeFile)
	for{
		buffer := make([]byte,8)
		_,err := reader.Read(buffer) //reader 按buffer的大小往buffer里面装内容,并且每read一次指针发生了对应的偏移
		if err == io.EOF{
			break
		}else{
			bwrite.Write(buffer)
		}
	}
	//让没有满缓存的内容也全部写到文件里面去
	bwrite.Flush()

	return nil

}