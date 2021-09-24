package main

import (
	"github.com/mozillazg/go-pinyin"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Site struct {
	Title       string
	Url         string
	Favicon     string
	SearchWords string
}

type Sites struct {
	Category string
	Links    []Site
}

type Categories struct {
	Cgs []Sites `yaml:"index"`
}

func getFavicon(url string) string {
	re := regexp.MustCompile("(?:\\w+\\.)+\\w+")
	host := re.FindString(url)
	//googleDownSrv := "https://www.google.com/s2/favicons?domain_url="
	yandexDownSrv := "http://favicon.yandex.net/favicon/"
	//ddgDownSrv := "https://icons.duckduckgo.com/ip3/www.google.com.ico"
	if len(host) > 0 {
		fav := "content/img/" + host + ".png"
		resp, err := http.Get(yandexDownSrv + host)
		if err != nil {
			return ""
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				return
			}
		}(resp.Body)

		out, err := os.Create(fav)
		if err != nil {
			return ""
		}
		defer func(out *os.File) {
			err := out.Close()
			if err != nil {
				return
			}
		}(out)

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return ""
		}
		fi, err := os.Stat(fav)
		if err != nil {
			return ""
		}
		if fi.Size() < 80 {
			return ""
		}

		return "img/" + host + ".png"
	} else {
		return ""
	}
}

func processSite(site *Site, force bool) *Site {
	if force {
		site.Favicon = getFavicon(site.Url)
	} else {
		file := strings.Replace(site.Favicon, "img/", "content/img-old/", 1)
		fi, err := os.Stat(file)
		if err == nil && fi.Size() > 80 {
			err = os.Rename(file, "content/"+site.Favicon)
			if err != nil {
				return nil
			}
		} else {
			site.Favicon = getFavicon(site.Url)
		}
	}

	site.SearchWords = site.Title
	a := pinyin.NewArgs()
	full := pinyin.Pinyin(site.Title, a)
	tmpstr := ""
	for _, v := range full {
		tmpstr += v[0]
	}
	site.SearchWords += " " + tmpstr
	a.Style = pinyin.FirstLetter
	first := pinyin.Pinyin(site.Title, a)
	tmpstr = ""
	for _, v := range first {
		tmpstr += v[0]
	}
	site.SearchWords += " " + tmpstr
	re := regexp.MustCompile("(?:\\w+\\.)+\\w+")
	host := re.FindString(site.Url)
	strs := strings.Split(host, ".")
	site.SearchWords += " " + strings.Join(strs[len(strs)-2:], ".")
	return site
}

func main() {
	err := os.Rename("content/img", "content/img-old")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Mkdir("content/img", 0755)
	if err != nil {
		log.Fatal(err)
	}

	ymlFile, err := ioutil.ReadFile("data/websites.yml")
	if err != nil {
		log.Fatal(err)
	}
	var data Categories
	err = yaml.Unmarshal(ymlFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	args := os.Args[1:]
	force := false
	if len(args) > 0 && args[0] == "--force" {
		force = true
	}
	for cg := range data.Cgs {
		for site := range data.Cgs[cg].Links {
			processSite(&data.Cgs[cg].Links[site], force)
		}
	}

	d, err := yaml.Marshal(&data)
	err = ioutil.WriteFile("data/new.yml", d, 0644)
	if err != nil {
		log.Fatal(err)
	}

	imgdir, err := os.Open("content/img")
	if err != nil {
		return
	}
	files, _ := imgdir.ReadDir(-1)
	_ = imgdir.Close()
	for _, file := range files {
		fi, _ := os.Stat("content/img/" + file.Name())
		if fi.Size() < 80 {
			_ = os.Remove("content/img/" + file.Name())
		}
	}
	//_ = os.Remove("content/img-old")
}
