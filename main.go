package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type reference struct {
	url string
	title string
}
func main() {
	var (
		filePath string
		outPutPath string
		proxy string
		references []reference
	)

	flag.StringVar(&filePath, "i", "./input.ref", "文件路径")
	flag.StringVar(&outPutPath, "o", "./reference.md", "输出文件路径")
	flag.StringVar(&proxy, "p", "", "代理配置")

	flag.Parse()
	err := ReadLine(filePath, Print, &references, proxy)

	fWrite, err2 := os.Create(outPutPath)
	if err2 != nil {
		panic(err2.Error())
	}

	for _, value := range references {
		fWrite.WriteString(fmt.Sprintf("- [ ] [%s](%s)\n", value.title, value.url))
	}

	if err != nil {
		return
	}
}

func Print(line string, proxy string) reference {

	req, err := http.NewRequest(http.MethodGet, line, nil)
	tr := &http.Transport{TLSClientConfig: &tls.Config{
		InsecureSkipVerify: true,
	}}
	if proxy != "" {
		proxyUrl, err := url.Parse(proxy)
		if err == nil { // 使用传入代理
			tr.Proxy = http.ProxyURL(proxyUrl)
		}
	}
	r, err := (&http.Client{Transport: tr}).Do(req)

	if err != nil {
		return reference{
			"",
			"",
		}
	}

	if title, ok := GetHtmlTitle(r.Body); ok {
		return reference{line, title}
	} else {
		println("Fail to get HTML title")
		return reference{
			"",
			"",
		}
	}
}

func ReadLine(fileName string, handler func(string, string) reference, references *[]reference, proxy string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	buf := bufio.NewReader(f)
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		*references = append(*references, handler(line, proxy))
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func isTitleElement(n *html.Node) bool {
	return n.Type == html.ElementNode && n.Data == "title"
}

func traverse(n *html.Node) (string, bool) {
	if isTitleElement(n) {
		return n.FirstChild.Data, true
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		result, ok := traverse(c)
		if ok {
			return result, ok
		}
	}

	return "", false
}

func GetHtmlTitle(r io.Reader) (string, bool) {
	doc, err := html.Parse(r)
	if err != nil {
		panic("Fail to parse html")
	}

	return traverse(doc)
}
