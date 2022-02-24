package main

import (
	"bytes"
	"encoding/base64"
	"github.com/mat/besticon/ico"
	"github.com/mozillazg/go-pinyin"
	"golang.org/x/image/draw"
	"gopkg.in/yaml.v2"
	"image"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

type Site struct {
	Title        string `yaml:"title"`
	Url          string `yaml:"url"`
	Favicon      string `yaml:"favicon"`
	SearchWords  string `yaml:"searchwords"`
	DataUriClass string `yaml:"datauriclass,omitempty"`
}

type Sites struct {
	Category string `yaml:"category"`
	Links    []Site `yaml:"links"`
}

type Categories struct {
	Cgs []Sites `yaml:"index"`
}

func getFavicon(url string) string {
	re := regexp.MustCompile("(?:[\\w-]+\\.)+\\w+")
	host := re.FindString(url)
	//googleDownSrv := "https://www.google.com/s2/favicons?domain_url="
	yandexDownSrv := "http://favicon.yandex.net/favicon/"
	//ddgDownSrv := "https://icons.duckduckgo.com/ip3/www.google.com.ico"
	if len(host) > 0 {
		fav := "content/img/" + host + ".png"
		time.Sleep(500 * time.Microsecond)
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

func generateDataUri(file string) string {
	if file == "" {
		return "nofavicon"
	} else {
		class := strings.Replace(file, "img/", "", 1)
		class = strings.Replace(class, ".png", "", 1)
		class = strings.Replace(class, ".ico", "", 1)
		class = strings.ReplaceAll(class, ".", "")
		if strings.Contains("0123456789", class[0:1]) {
			class = "c" + class
		}
		input, _ := os.Open("content/" + file)
		defer func(input *os.File) {
			err := input.Close()
			if err != nil {
				log.Fatal("converting image to data uri failed")
			}
		}(input)
		out := new(bytes.Buffer)
		var src image.Image
		if strings.Contains(file, ".ico") {
			src, _ = ico.Decode(input)
		} else {
			// Decode the image (from PNG to image.Image):
			src, _ = png.Decode(input)
		}
		// Set the expected size that you want:
		dst := image.NewRGBA(image.Rect(0, 0, 16, 16))
		// Resize:
		draw.BiLinear.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)

		// Encode to `output`:
		err := png.Encode(out, dst)
		if err != nil {
			return ""
		}

		base64Img := base64.StdEncoding.EncodeToString(out.Bytes())
		uri := "data:image/png;base64," + base64Img
		css := []byte("." + class + " {\n" + "  background-image: url(\"" + uri + "\");\n  padding: 16px 0 0 0;\n}\n")
		cssfile, err := os.OpenFile("static/assets/siteimg.css", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("failed creating file: %s", err)
		}
		filestat, _ := cssfile.Stat()
		if filestat.Size() == 0 {
			nofav := []byte(".nofavicon {\n  background-image: url(\"data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABAAAAAQCAYAAAAf8/9hAAABs0lEQVR4AWL4//8/RRjO8Iucx+noO0O2qmlbUEnt5r3Juas+hsQD6KaG7dqCKPgx72Pe9GIY27btZBrbtm3btm0nO12D7tVXe63jqtqqU/iDw9K58sEruKkngH0DBljOE+T/qqx/Ln718RZOFasxyd3XRbWzlFMxRbgOTx9QWFzHtZlD+aqLb108sOAIAai6+NbHW7lUHaZkDFJt+wp1DG7R1d0b7Z88EOL08oXwjokcOvvUxYMjBFCamWP5KjKBjKOpZx2HEPj+Ieod26U+dpg6lK2CIwTQH0oECGT5eHj+IgSueJ5fPaPg6PZrz6DGHiGAISE7QPrIvIKVrSvCe2DNHSsehIDatOBna/+OEOgTQE6WAy1AAFiVcf6PhgCGxEvlA9QngLlAQCkLsNWhBZIDz/zg4ggmjHfYxoPGEMPZECW+zjwmFk6Ih194y7VHYGOPvEYlTAJlQwI4MEhgTOzZGiNalRpGgsOYFw5lEfTKybgfBtmuTNdI3MrOTAQmYf/DNcAwDeycVjROgZFt18gMso6V5Z8JpcEk2LPKpOAH0/4bKMCAYnuqm7cHOGHJTBRhAEJN9d/t5zCxAAAAAElFTkSuQmCC\");\n  padding: 16px 0 0 0;\n}\n")
			_, err := cssfile.Write(nofav)
			if err != nil {
				return ""
			}
		}
		_, err = cssfile.Write(css)
		if err != nil {
			return ""
		}
		err = cssfile.Close()
		if err != nil {
			return ""
		}
		return class
	}
}

func processSite(site *Site, force bool) *Site {
	if force {
		site.Favicon = getFavicon(site.Url)
	} else {
		fileold := strings.Replace(site.Favicon, "img/", "content/img-old/", 1)
		fiold, err := os.Stat(fileold)
		fi, err2 := os.Stat("content/" + site.Favicon)
		if err == nil && fiold.Size() > 80 {
			err = os.Rename(fileold, "content/"+site.Favicon)
			if err != nil {
				return nil
			}
		} else {
			if err2 != nil || fi.Size() < 80 {
				site.Favicon = getFavicon(site.Url)
			}
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
	re := regexp.MustCompile("(?:[\\w-]+\\.)+\\w+")
	host := re.FindString(site.Url)
	// strs := strings.Split(host, ".")
	// site.SearchWords += " " + strings.Join(strs[len(strs)-2:], ".")
	site.SearchWords += " " + host
	site.DataUriClass = generateDataUri(site.Favicon)
	return site
}

func main() {
	err := os.Rename("content/img", "content/img-old")
	if err != nil {
		log.Fatal(err)
	}
	err = os.Rename("static/assets/siteimg.css", "static/assets/siteimg-old.css")
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
