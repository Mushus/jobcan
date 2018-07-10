package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"time"
)

var (
	now = time.Now()
)
var (
	date   = flag.String("date", "", fmt.Sprintf("打刻日時 format: 2006-01-02, default: %s", now.Format("2006-01-02")))
	ftime  = flag.String("time", "", fmt.Sprintf("打刻時間 format: 15:04, default: %s", now.Format("15:04")))
	code   = flag.String("code", "", "ログインのURLに含まれるcode値")
	lat    = flag.String("lat", "", "緯度")
	lon    = flag.String("lon", "", "経度")
	gid    = flag.String("gid", "", "グループID")
	yakin  = flag.Bool("yakin", false, "夜勤モード default: false")
	taikin = flag.Bool("taikin", false, "退勤フラグ default: false")
	reason = flag.String("reason", "", "理由")
)

var (
	tokenRegexp = regexp.MustCompile("<input\\stype=\"hidden\"\\sclass=\"token\"\\sname=\"token\"\\svalue=\"(\\w*)\">")
)

func main() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage of %s:
  %s [OPTIONS] ARGS...
  Options\n`, os.Args[0], os.Args[0])
		flag.PrintDefaults()
	}

	if *code != "" {
		fmt.Printf("code: %v\nlat: %v\nlon: %v\ngroup_id: %v\n", *code, *lat, *lon, *gid)
		sid, err := login(*code)
		if err != nil {
			log.Fatalf("failed to login: %#v", err)
		}
		fmt.Printf("session: %v\n", sid)
		token := ""
		for {
			var err error
			token, err = dakoku(sid, token)
			if err != nil {
				log.Fatalf("failed to login: %#v", err)
			}
			if token == "" {
				break
			}
		}
		fmt.Println("打刻!")
	} else {
		flag.Usage()
	}
}

func login(code string) (string, error) {
	q := url.Values{
		"code": {code},
	}

	uri := url.URL{
		Scheme:   "http",
		Host:     "jobcan.jp",
		Path:     "/m",
		RawQuery: q.Encode(),
	}

	resp, err := http.Get(uri.String())
	if err != nil {
		return "", fmt.Errorf("failed to login: %#v", err)
	}

	// NOTE: 何故かset-cookieが２つ帰ってくる
	rawCookie := resp.Header["Set-Cookie"][1]

	cookie, err := url.ParseQuery(rawCookie)
	if err != nil {
		return "", fmt.Errorf("failed to parse cookie: %#v", err)
	}

	sid := cookie.Get("sid")
	return sid, nil
}

func dakoku(sid string, token string) (string, error) {
	yakinPrm := []string{}
	if *yakin {
		yakinPrm = []string{"1"}
	}

	aditItemPrm := []string{"打刻"}
	if *taikin {
		aditItemPrm = []string{"退勤"}
	}

	q := url.Values{
		"lon":         {*lon},
		"lat":         {*lat},
		"year":        {now.Format("2006")},
		"month":       {now.Format("1")},
		"day":         {now.Format("2")},
		"reason":      {*reason}, // 理由
		"time":        {},        // ?
		"group_id":    {*gid},
		"position_id": {},          // ?
		"adit_item":   aditItemPrm, // 打刻, 退勤
		"yakin":       yakinPrm,    // 1: 夜勤
	}

	method := http.MethodGet
	urlPath := "/m/work/stamp-save-confirm/"
	if token != "" {
		method = http.MethodPost
		urlPath = "/m/work/stamp-save-smartphone/"
		q.Add("confirm", "はい")
		q.Add("token", token)
	}

	uri := url.URL{
		Scheme:   "https",
		Host:     "ssl.jobcan.jp",
		Path:     urlPath,
		RawQuery: q.Encode(),
	}

	req, err := http.NewRequest(method, uri.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %#v", err)
	}

	cookie := fmt.Sprintf("sid=%s", sid)
	req.Header.Set("cookie", cookie)

	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to dakoku request: %#v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to dakoku: %#v", err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to get dakoku body: %#v", err)
	}

	if token == "" {
		// NOTE: スクレイピングだとHTMLが壊れてたときにつらそう
		matchString := tokenRegexp.FindSubmatch(b)
		if len(matchString) >= 2 {
			return string(matchString[1]), nil
		}
		//fmt.Printf("%#v", matchString)
	}

	return "", nil
}
