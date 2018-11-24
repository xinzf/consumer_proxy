package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"github.com/satori/go.uuid"
	"io"
	"os"
)

func EncodeMD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

func FileMd5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}

func EncodeSha1(str string) string {
	s := sha1.New()
	s.Write([]byte(str))
	bs := s.Sum(nil)

	return hex.EncodeToString(bs)
}

func EncodeBase64(data []byte) string {
	encodeString := base64.StdEncoding.EncodeToString(data)
	return encodeString
}

func PathExists(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if os.IsNotExist(err) {
		return false
	}

	return false
}

func Mkdir(path string) error {
	exist := PathExists(path)
	if !exist {
		err := os.MkdirAll(path, os.ModePerm)
		return err
	}

	return nil
}

func GetIP() ([]string, error) {
	var ips []string
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ips, err
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips, nil
}

func Ip2num(ip string) int {
	canSplit := func(c rune) bool { return c == '.' }
	lisit := strings.FieldsFunc(ip, canSplit) //[58 215 20 30]
	//fmt.Println(lisit)
	ip1_str_int, _ := strconv.Atoi(lisit[0])
	ip2_str_int, _ := strconv.Atoi(lisit[1])
	ip3_str_int, _ := strconv.Atoi(lisit[2])
	ip4_str_int, _ := strconv.Atoi(lisit[3])
	return ip1_str_int<<24 | ip2_str_int<<16 | ip3_str_int<<8 | ip4_str_int
}

func Num2ip(num int) string {
	ip1_int := (num & 0xff000000) >> 24
	ip2_int := (num & 0x00ff0000) >> 16
	ip3_int := (num & 0x0000ff00) >> 8
	ip4_int := num & 0x000000ff
	//fmt.Println(ip1_int)
	data := fmt.Sprintf("%d.%d.%d.%d", ip1_int, ip2_int, ip3_int, ip4_int)
	return data
}

func Date() string {
	return TimeFormat(time.Now())
}

func TimeFormat(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

func GreenwichToDate(date string) string {
	pos := strings.IndexAny(date, "+")
	if pos != -1 {
		date = date[:pos]
	}

	pos = strings.IndexAny(date, "T")
	if pos != -1 {
		date = strings.Replace(date, "T", " ", pos)
	}
	return date
}

func UUID() string {
	id, _ := uuid.NewV4()
	return strings.Replace(id.String(), "-", "", -1)
}

func JsonEncode(o interface{}) (ret []byte, err error) {
	ret, err = json.Marshal(o)
	return
}
