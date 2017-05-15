package main

import (
	"net/http"
	"testing"
	"math/rand"
	"sync"
	"fmt"
	"time"
	"golang.org/x/net/proxy"
	"net"
	"github.com/jiusanzhou/tentacle/server"
)

func Socks5Client(addr string, auth ...*proxy.Auth) (client *http.Client, err error) {

	dialer, err := proxy.SOCKS5("tcp", addr,
		nil,
		&net.Dialer {
			Timeout: 30 * time.Second,
			KeepAlive: 30 * time.Second,
		},
	)
	if err != nil {
		return
	}

	transport := &http.Transport{
		Proxy: nil,
		Dial: dialer.Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	client = &http.Client { Transport: transport }

	return
}


var (
	html_urls = []string{
		"https://www.baidu.com",
		"http://cn.bing.com",
		"http://www.sina.com.cn/",
		"http://www.qq.com/",
	}

	img_urls = []string{
		"https://image1.ljcdn.com/neirong-image/neirong1489544691phpnnMJKB.png",
	}

	binary_urls = []string{
		"https://xianzhi.aliyun.com/forum/attachment/big_size/wafbypass_sql.pdf",
	}
)

func get(client *http.Client, url string) {
	resp, err := client.Get(url)
	if err != nil {
		resp.Body.Close()
	}
	fmt.Printf("finished %s\n", url)
}

func run(){
	server.Main()
}

func TestAll(t *testing.T) {

}

func BenchmarkHtml(b *testing.B) {
	// go run()
	// time.Sleep(5*time.Second)
	wg := sync.WaitGroup{}
	for i := 0; i<b.N; i++ {
		wg.Add(1)
		client, _ := Socks5Client("127.0.0.1:8888")
		fmt.Println("start a goroutin")
		go func(){
			for n:=0; n<b.N; n++{
				get(client, html_urls[rand.Intn(len(html_urls))])
			}
			wg.Done()
		}()
	}
	fmt.Println("wait all finished")
	wg.Wait()
}

func BenchmarkImg(b *testing.B) {
	wg := sync.WaitGroup{}
	for i := 0; i<b.N; i++ {
		wg.Add(1)
		client, _ := Socks5Client("127.0.0.1:8888")
		fmt.Println("start a goroutin")
		go func(){
			for n:=0; n<b.N; n++{
				get(client, img_urls[rand.Intn(len(img_urls))])
			}
			wg.Done()
		}()
	}
	fmt.Println("wait all finished")
	wg.Wait()
}

func BenchmarkBinary(b *testing.B) {
	wg := sync.WaitGroup{}
	for i := 0; i<b.N; i++ {
		wg.Add(1)
		client, _ := Socks5Client("127.0.0.1:8888")
		fmt.Println("start a goroutin")
		go func(){
			for n:=0; n<b.N; n++{
				get(client, binary_urls[rand.Intn(len(binary_urls))])
			}
			wg.Done()
		}()
	}
	fmt.Println("wait all finished")
	wg.Wait()
}
