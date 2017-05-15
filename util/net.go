package util

import (
	"net"
	"net/http"
	"time"
	"crypto/tls"
	"fmt"
	"encoding/json"
	"io/ioutil"
)

// get mac address from target machine.
// may be machine has multi interfaces,
// we only get the first one.
func GetMacAddr()(macStr string){
	ifs, _ := net.Interfaces()
	for _, i := range ifs{
		macStr = i.HardwareAddr.String()
		if macStr != ""{
			break
		}
	}
	return
}

const (
	defaultTimeOut = 20 * time.Second
	defaultUserAgent = "Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/50.0.2661.102 Safari/537.36"
	TestOcasEndpoint = "https://api.fast.com/netflix/speedtest"
	TestOcasToken = "YXNkZmFzZGxmbnNkYWZoYXNkZmhrYWxm"
)

var (
	req *http.Request //request
	resp *http.Response // response
	t int64 // time unix second
	n int // response content length
)

type speedObject struct{
	size int64
	coast int64
	speed int64
}

// test the internet speed.
// api from https://fast.com
// end point and token should get from index and inner js code,
// but I just get out from https://fast.com/app-53b06a.js
func GetSpeedIn()(speed int64){

	fmt.Println("This will coast lots of time, please wait for a while.\nDon't worry if nothing print on screen.")

	client := &http.Client{
		// Timeout: defaultTimeOut,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}

	baseUrl := fmt.Sprintf("%s?token=%s&urlCount=5&https=true", TestOcasEndpoint, TestOcasToken)
	c, _, err := downloadSource(client, baseUrl)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}

	x := []map[string]string{}
	err = json.Unmarshal(c, &x)
	if err !=nil {
		fmt.Printf("get end point error, %s\n.", err.Error())
		return
	}

	scs := []speedObject{}
	fmt.Printf("total we has %d objects to request.\n", len(x))
	for i, d := range x{
		if d["url"] == "" {
			fmt.Printf("the %d url is empty.\n", i)
			continue
		}

		t = time.Now().Unix()
		r, _, err := downloadSource(client, d["url"])
		n = len(r)
		size := int64(n)
		coast := time.Now().Unix()-t
		scs = append(scs, speedObject{size, coast, size/coast})
		// fmt.Printf("size: %d, coast: %d, speed: %d.\n", size, coast, 8*size/coast/1024/1024)
		if err!=nil{
			fmt.Printf("the %d request with error, %s.\n", i, err.Error())
			continue
		}
	}

	var total int64
	for _, i:=range scs {
		total += 8*i.speed/1024
	}
	speed = total/int64(len(scs))
	return
}

// download content from source url.
func downloadSource(client *http.Client, url string)(res []byte, n int, err error){
	req, _ = http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", defaultUserAgent)
	resp, err = client.Do(req)
	defer resp.Body.Close()
	res, err = ioutil.ReadAll(resp.Body)
	return
}