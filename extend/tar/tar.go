package tar

import (
	"archive/tar"
	"github.com/rs/zerolog/log"
	"io"
	"os"
)

// tar包实现了文件的打包功能,可以将多个文件或者目录存储到单一的.tar压缩文件中
// tar本身不具有压缩功能,只能打包文件或目录
//熟悉以上便知道它能做的是将多个文件内容不压缩的在一个文件上具备

//zipFile 生成的打包文件全路径；filesPath需要打包的文件路径切片
//归纳成3步：1.创建合并包的file对象；2：根据file对象生成writer对象，3.每一个文件写入write都需要2步，第一写header，第二写body
func TagFiles(tarFilePath string,filesPaths ...string)(b bool,error error){

	fw,err :=os.Create(tarFilePath)//生成带要写入的合并的包
	defer fw.Close()

	if err != nil {
		log.Error().Msg(err.Error())
		error = err
		return
	}

	// 通过fw创建一个tar.Writer
	tw :=tar.NewWriter(fw)

	defer func(){
		if err:=tw.Close();err!=nil{
			log.Error().Msg(err.Error())
			error = err
			b = false //这一步涉及到将缓冲区的内容写到文件路径，所以失败就应该改变返回的结果
		}
	}()

	for _,filePath := range filesPaths{

		sourceFileInfo,err := os.Stat(filePath)//为写header  取得一个文件信息 ，写header的时候便可在此取

		hdr,err :=tar.FileInfoHeader(sourceFileInfo,"")//tar 从fileinfo中取 header的信息，并返回

		if err :=tw.WriteHeader(hdr);err!=nil{  //writer对象将header信息写进自己里面来
			log.Error().Msg(err.Error())
			error = err
			return
		}
		fl,err :=os.Open(filePath)
		//通过文件路径获取文件资源
		if err!=nil{
			log.Error().Msg(err.Error())
			error = err
			return
		}

		if _,err :=io.Copy(tw,fl);err!=nil{
			log.Error().Msg(err.Error())
			error = err
			return
		}
		fl.Close()
	}
	b = true
	return
}