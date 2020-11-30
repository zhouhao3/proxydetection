package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func run(context *cli.Context) error {
	fmt.Println("===========Start================")

	if len(context.Args()) != 2 {
		fmt.Printf("Incorrect Usage.\n\n")
		cli.ShowCommandHelp(context, "")
		return errors.New("proxydetection requires exactly 2 arguments")
	}

	proxy := context.Args()[0]
	filePath := context.Args()[1]
	fileType := context.String("file-type")
	specialURLPath := context.String("special-path")
	resultFileName := "result_url"

	var result2 []string

	if fileType != "glider-log" && fileType != "url-file" {
		return errors.New("file-type only supports glider-log or url-file")
	}

	result, err := read(filePath)
	if err != nil {
		return err
	}

	if fileType == "glider-log" {
		result = getURL(result)
	}

	sort.Strings(result)
	result = removeDuplicates(result)

	if specialURLPath != "" {
		specialUrls, err := read(specialURLPath)
		if err != nil {
			return err
		}

		result = removeSpecialURL(result, specialUrls)
	}

	for _, url := range result {
		_, status := urlTest(url, proxy)
		if status == 200 {
			continue
		} else {
			result2 = append(result2, url)
		}
	}

	result2 = mergeURL(result2)

	err = write(result2, resultFileName)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Printf("The result has been written to file %s\n", resultFileName)
	fmt.Println("===========Finshed==============")

	return nil
}

func mergeURL(a []string) (ret []string) {
	contain := false
	for i := 0; i < len(a)-1; i++ {
		t := strings.SplitAfterN(a[i], ".", 2)
		s := []rune(t[0])
		if len(s) > 2 {
			s1 := s[:len(s)-2]
			contain = false
			for j := i + 1; j < len(a)-1; j++ {
				if strings.HasSuffix(a[j], t[1]) {
					if strings.HasPrefix(a[j], string(s1)) {
						a = append(a[:j], a[j+1:]...)
						j = j - 1
						contain = true
					}
				}
			}
			if contain {
				s[len(s)-2] = '*'
				t[0] = string(s)
			}
			a[i] = t[0] + t[1]
		}

		ret = append(ret, a[i])
	}

	return
}

func removeSpecialURL(a, b []string) (ret []string) {
	var repeat bool
	for _, s := range a {
		for _, ss := range b {
			if s == ss {
				repeat = true
				break
			}
		}
		if repeat == true {
			repeat = false
			continue
		}

		ret = append(ret, s)
	}

	return
}

func write(str []string, f string) error {
	file, err := os.Create(f)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, s := range str {
		file.WriteString(s + "\n")
	}

	return nil
}

func read(f string) (urls []string, err error) {
	file, err := os.Open(f)
	if err != nil {
		return
	}
	defer file.Close()

	fileScanner := bufio.NewScanner(file)

	for {
		if !fileScanner.Scan() {
			break
		}
		line := fileScanner.Text()
		line = strings.TrimSpace(line)

		urls = append(urls, line)
	}
	return
}

func getURL(a []string) (ret []string) {
	for _, url := range a {
		strSlice := strings.Split(url, " ")
		if len(strSlice) < 10 {
			continue
		}
		strURL := strings.Split(strSlice[6], ":")
		if len(strURL) < 2 {
			continue
		}
		if len(strings.Split(strURL[0], ".")) != 3 {
			continue
		}

		ret = append(ret, strURL[0])
	}

	return
}

func removeDuplicates(a []string) (result []string) {
	l := len(a)

	for i := 0; i < l; i++ {
		if i > 0 && a[i-1] == a[i] {
			continue
		}

		result = append(result, a[i])
	}

	return
}

func urlTest(urlAddr, proxyAddr string) (Speed, Status int) {
	urlAddr = "http://" + urlAddr
	proxy, err := url.Parse(proxyAddr)
	if err != nil {
		fmt.Println(err)
		return
	}

	netTransport := &http.Transport{
		Proxy:                 http.ProxyURL(proxy),
		MaxConnsPerHost:       10,
		ResponseHeaderTimeout: time.Second * time.Duration(5),
	}

	httpClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	begin := time.Now()
	res, err := httpClient.Get(urlAddr)
	if err != nil {
		return
	}
	defer res.Body.Close()

	speed := int(time.Now().Sub(begin).Nanoseconds() / 1000 / 1000)

	if res.StatusCode != http.StatusOK {
		return
	}

	return speed, res.StatusCode
}

func main() {
	app := cli.NewApp()
	app.Name = "proxydetection"
	app.Version = "0.1"
	app.Usage = "Check whether the specified proxy can connect to the URL in the log file"
	app.UsageText = "proxydetection [global options] PROXY FILE"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "file-type",
			Value: "glider-log",
			Usage: "Specified file type (glider-log or url-file).The default value is glider-log",
		},
		cli.StringFlag{
			Name:  "special-path",
			Usage: "Contains the path of the evil file with a special URL. The URL in this file will be ignored and will not be tested with proxy",
		},
	}

	app.Action = run
	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}
