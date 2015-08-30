package main

import (
	"encoding/json"
	"fmt"
	// "os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/PuerkitoBio/goquery"
	"weibo.com/golang-util/api"
)

type APIReader struct {
	HtmlCode string `json:"htmlCode"`
	Total    string `json:"total"`
	Pager    string `json:"pager"`
}

func (r *APIReader) Reset() {
	r.HtmlCode = ""
	r.Total = ""
	r.Pager = ""
}

func main() {
	// GetPage("alternate-heel-touchers", "http://www.bodybuilding.com/exercises/detail/view/name/alternate-heel-touchers")
	// GetPage("alternate-heel-touchers", "http://www.bodybuilding.com/exercises/detail/view/name/double-kettlebell-push-press")
	list := GetPageList()
	for name, url := range list {
		GetPage(name, url)
		time.Sleep(time.Second * 10)
	}
}

func GetPage(name, url string) {
	exerciseData := make(map[string]string)
	doc, _ := goquery.NewDocument(url)

	//抓取内容
	doc.Find("#exerciseDetails span.row").Each(func(i int, s *goquery.Selection) {
		node := explode(s.Text())
		if 1 < len(node) {
		}
		exerciseData[trim(node[0])] = trim(node[1])
	})

	//抓取分数
	rate := doc.Find("#exerciseRating span.rating").Text()
	exerciseData["rate"] = rate

	//抓取视频
	if vedio, ok := doc.Find("#videoContainer #maleVideo source").Attr("src"); ok {
		exerciseData["vedio"] = vedio
	}

	//抓取图片
	pics := make([]string, 0)
	doc.Find("div.photoLeft a img").Each(func(i int, s *goquery.Selection) {
		if str, ok := s.Attr("src"); ok {
			pics = append(pics, trim(str))
		}
	})
	doc.Find("div.photoRight a img").Each(func(i int, s *goquery.Selection) {
		if str, ok := s.Attr("src"); ok {
			pics = append(pics, trim(str))
		}
	})

	picsJson, _ := json.Marshal(pics)

	//抓取相关动作
	extra := make([]string, 0)
	doc.Find("#altExerciseCon div.exerciseName h3 a").Each(func(i int, s *goquery.Selection) {
		extra = append(extra, trim(s.Text()))
	})

	extraJson, _ := json.Marshal(extra)

	fmt.Printf("%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s\n", name, exerciseData["vedio"], exerciseData["rate"], exerciseData["Type"], exerciseData["Main Muscle Worked"], exerciseData["Other Muscles"], exerciseData["Equipment"], exerciseData["Mechanics Type"], exerciseData["Level"], exerciseData["Sport"], exerciseData["Force"], string(picsJson), string(extraJson))
	// fmt.Println(name, exerciseData["vedio"], exerciseData["rate"], exerciseData["Type"], exerciseData["Main Muscle Worked"], exerciseData["Other Muscles"], exerciseData["Equipment"], exerciseData["Mechanics Type"], exerciseData["Level"], exerciseData["Sport"], exerciseData["Force"], string(picsJson), string(extraJson))

}

func GetPageList() map[string]string {
	list := make(map[string]string)
	Regexp, err := regexp.Compile(`http://www.bodybuilding.com/exercises/detail/view/name/(.*)`)
	if nil != err {
		fmt.Println("正则错误")
		return nil
	}

	var page, total int
	var since_id string
	for page == 0 || page != total {
		since_id = strconv.Itoa(page)
		reader := new(APIReader)
		api.PostRequest(reader, "http://www.bodybuilding.com/exercises/ajax/getfinderdata/", "orderByField=exerciseName&orderByDirection=ASC", map[string]string{"params": "muscleID=13,3,18,5,17,4,15,6,9,7,1,12,2,11,14,10,8;exerciseTypeID=2,6,4,7,1,3,5;equipmentID=9,14,2,10,5,6,4,15,1,8,11,3,7;mechanicTypeID=1,2,11;force=Push,Pull,Static,N/A;sport=Yes,No;levelID=1,3,2", "page": since_id})
		nodeReader := strings.NewReader(reader.HtmlCode)
		total, _ = strconv.Atoi(reader.Total)
		if page+15 < total {
			page += 15
		} else {
			page = total
		}

		doc, err := goquery.NewDocumentFromReader(nodeReader)
		if nil != err {
			fmt.Println(err)
			return nil
		}
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			if url, ok := s.Attr("href"); ok {
				str := Regexp.FindAllStringSubmatch(url, 10)
				if 0 != len(str) {
					list[str[0][1]] = str[0][0]
				}
			}
		})
		time.Sleep(time.Second * 2)
		// fmt.Println(len(list), list)
		// if 45 == page {
		// 	return list
		// }
	}
	return list
}

func explode(str string) []string {
	return strings.FieldsFunc(str, func(c rune) bool {
		if ':' == c || '：' == c {
			return true
		} else {
			return false
		}
	})
}

func trim(str string) string {
	return strings.TrimFunc(str, func(c rune) bool {
		return !unicode.IsLetter(c) && !unicode.IsNumber(c)
	})
}
