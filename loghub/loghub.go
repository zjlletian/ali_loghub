package loghub

import (
	"bytes"
	"compress/zlib"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

//loghub配置
type Config struct {
	AccessKey    string // 用于loghub的AccessKeyId
	AccessSecret string // 用于loghub的AccessKeySecret
	EndPoint     string // loghub服务器地址
	LogStore     string // loghub的LogStore名称
}

//发送日志
func SendLog(config Config, logs []*Log) (err error) {
	body := logToPB(logs)
	bodylen := fmt.Sprintf("%d", len(body))
	var buf bytes.Buffer
	compressor := zlib.NewWriter(&buf)
	compressor.Write(body)
	compressor.Close()
	zlibbody := buf.Bytes()
	zliblen := fmt.Sprintf("%d", len(zlibbody))
	md5str := strings.ToUpper(fmt.Sprintf("%x", md5.Sum(zlibbody)))

	date := time.Now().UTC().Format(time.RFC1123)
	logheaders := "x-log-apiversion:0.6.0\nx-log-bodyrawsize:" + bodylen + "\nx-log-compresstype:deflate\nx-log-signaturemethod:hmac-sha1"
	path := "/logstores/" + config.LogStore + "/shards/lb"
	singstring := querySign(config.AccessSecret, "POST", md5str, "application/x-protobuf", date, logheaders, path)

	headers := make(map[string]string)
	headers["Date"] = date
	headers["Content-Type"] = "application/x-protobuf"
	headers["Content-MD5"] = md5str
	headers["Content-Length"] = zliblen
	headers["x-log-apiversion"] = "0.6.0"
	headers["x-log-bodyrawsize"] = bodylen
	headers["x-log-compresstype"] = "deflate"
	headers["x-log-signaturemethod"] = "hmac-sha1"
	headers["Authorization"] = "LOG " + config.AccessKey + ":" + singstring
	resp, code, err := doHttpRequest("POST", "http://"+config.EndPoint+path, headers, zlibbody)
	if err != nil {
		return
	}
	if code != 200 {
		err = errors.New(string(resp))
	}
	return
}

//将日志转换为pb格式
func logToPB(logs []*Log) (buffers []byte) {
	group := &LogGroup{
		Logs: logs,
	}
	buffer, err := proto.Marshal(group)
	if err != nil {
		fmt.Println(err)
	}
	buffers = buffer
	return
}

//请求签名
func querySign(key string, method string, md5str string, contenttype string, date string, logheaders string, path string) (sign string) {
	signString := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s", method, md5str, contenttype, date, logheaders, path)
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(signString))
	sign = base64.StdEncoding.EncodeToString(mac.Sum(nil))
	return
}

//发送请求
func doHttpRequest(method string, url string, header map[string]string, body []byte) (respbody []byte, statusCode int, err error) {
	client := &http.Client{}
	req, e := http.NewRequest(method, url, bytes.NewReader(body))
	if e != nil {
		err = e
		return
	}
	for key, value := range header {
		req.Header.Set(key, value)
	}
	resp, e := client.Do(req)
	if e != nil {
		err = e
		return
	}
	defer resp.Body.Close()
	statusCode = resp.StatusCode
	respbody, err = ioutil.ReadAll(resp.Body)
	return
}
