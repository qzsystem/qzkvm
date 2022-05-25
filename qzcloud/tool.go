package qzcloud

import (
	"bytes"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"html/template"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func isDirExits(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func mkdir(path string) error {
	er := isDirExits(path)
	if er==false {
		err := os.MkdirAll(path, os.ModePerm)
		return err
	}
	return nil
}
func RemoveFile(path string)error{
	path1:=strings.ReplaceAll(path,"\\","/")
	path1 = strings.ReplaceAll(path1,"//","/")
	path1 = strings.ReplaceAll(path1,"//","/")

	if path1 =="/"{
		return nil
	}
	if path1 ==""{
		return nil
	}
	pathSlice:=strings.Split(path1,"/")
	if len(pathSlice)<2{
		return nil
	}
	if path1 =="/root"{
		return nil
	}
	if path1 =="/root/"{
		return nil
	}

	err:=os.RemoveAll(path)
	return err
}

func Absum(param ...string)string{
	var count int64 =1
	for _,val:=range param {
		intval,err:=strconv.ParseInt(val,10,64)
		if err!=nil{
			intval=0
		}
		count=count*intval
	}

	d:=strconv.FormatInt(count,10)
	return  d
}

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func unescaped (content string) interface{} { return template.HTML(content) }

func formatFloat(value float64) float64 {
	value,_=strconv.ParseFloat(fmt.Sprintf("%.2f",value),64)
	return value

}

func GetUUIDBuild() string {
	u, _ := uuid.NewV4()
	return u.String()
}

func Exec_shell(shellstring string) (string, error) {
	//函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
	cmd := exec.Command("/bin/bash", "-c", shellstring)

	//读取io.Writer类型的cmd.Stdout，再通过bytes.Buffer(缓冲byte类型的缓冲器)将byte类型转化为string类型(out.String():这是bytes类型提供的接口)
	var out bytes.Buffer
	cmd.Stdout = &out

	//Run执行c包含的命令，并阻塞直到完成。  这里stdout被取出，cmd.Wait()无法正确获取stdin,stdout,stderr，则阻塞在那了
	err := cmd.Run()

	return out.String(), err
}