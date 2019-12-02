package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

type config struct {
	Path        string `json:"path"`
	Date        string `json:"date"`
	Feed        string `json:"feed"`
	TorrentPath string `json:"torrentPath"`
}

func downloadTorrentFiles(url string, localPath string) error {

	output, err := os.Create(localPath)
	if err != nil {
		return err
	}

	defer output.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(output, resp.Body)
	return err

}

func readConfig() (date time.Time, feed string, torrentPath string, err error) {

	var confPath string = os.Args[1]

	file, err := os.Open(confPath)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	byteValue, _ := ioutil.ReadAll(file)
	file.Close()
	var conf config
	json.Unmarshal(byteValue, &conf)
	date, err = time.Parse("02 Jan 2006 15:04 -0700", conf.Date)
	fmt.Println(conf.Date)
	return date, conf.Feed, conf.TorrentPath, err
}

func writeConfig(time time.Time, url string, torrentPath string) {

	var confPath string = os.Args[1]

	conf := config{
		Path:        confPath,
		Date:        time.Format("02 Jan 2006 15:04 -0700"),
		Feed:        url,
		TorrentPath: torrentPath,
	}

	file, _ := json.MarshalIndent(conf, "", " ")
	_ = ioutil.WriteFile(confPath, file, 0644)
}

func main() {
	fp := gofeed.NewParser()
	lastDownloadDate, url, torrentPath, err := readConfig()
	feed, _ := fp.ParseURL(url)

	if err != nil {
		fmt.Println("Unable to Parse Config: \n", err)
		os.Exit(1)
	}
	for i := 0; i < len(feed.Items); i++ {
		t, _ := time.Parse("Mon, 02 Jan 2006 15:04:05 -0700", feed.Items[i].Published)
		if t.After(lastDownloadDate) {
			fmt.Println("Downloaded: ", feed.Items[i].Link)
			fmt.Println(feed.Items[i].Title)
			title := strings.Split(feed.Items[i].Title, "]")
			go downloadTorrentFiles(feed.Items[i].Link, torrentPath+strings.Trim(title[3][3:], " ")+".torrent")
			lastDownloadDate = t
		}
	}
	writeConfig(lastDownloadDate, url, torrentPath)

}
