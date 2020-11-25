package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/wcharczuk/go-chart/v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"time"
)
import "github.com/gocolly/colly"

const (
	URL = "http://api.fund.eastmoney.com/f10/lsjz?fundCode=%s&pageIndex=0&pageSize=%d"

)
type JSONTime time.Time
func (t *JSONTime) UnmarshalJSON(data []byte) (err error) {
	now, err := time.ParseInLocation(`"2006-01-02"`, string(data), time.Local)
	*t = JSONTime(now)
	return
}

func getFont(fontFile string) *truetype.Font {
	// 读字体数据
	fontBytes, err := ioutil.ReadFile(fontFile)
	if err != nil {
		log.Println(err)
		return nil
	}
	font, err := truetype.Parse(fontBytes)
	if err != nil {
		log.Println(err)
		return nil
	}
	return font
}
type RespData struct {
	Data struct {
		LSJZList []struct {
			FSRQ      JSONTime      `json:"FSRQ"`
			DWJZ      interface{}      `json:"DWJZ"`
			LJJZ      interface{}      `json:"LJJZ"`
			SDATE     interface{} `json:"SDATE"`
			ACTUALSYI string      `json:"ACTUALSYI"`
			NAVTYPE   string      `json:"NAVTYPE"`
			JZZZL     string      `json:"JZZZL"`
			SGZT      string      `json:"SGZT"`
			SHZT      string      `json:"SHZT"`
			FHFCZ     string      `json:"FHFCZ"`
			FHFCBZ    string      `json:"FHFCBZ"`
			DTYPE     interface{} `json:"DTYPE"`
			FHSP      string      `json:"FHSP"`
		} `json:"LSJZList"`
		FundType  string      `json:"FundType"`
		SYType    interface{} `json:"SYType"`
		IsNewType bool        `json:"isNewType"`
		Feature   string      `json:"Feature"`
	} `json:"Data"`
	ErrCode    int         `json:"ErrCode"`
	ErrMsg     interface{} `json:"ErrMsg"`
	TotalCount int         `json:"TotalCount"`
	Expansion  interface{} `json:"Expansion"`
	PageSize   int         `json:"PageSize"`
	PageIndex  int         `json:"PageIndex"`
}

func FetchFundAllNetValue(code string) () {
	collector := colly.NewCollector()

	collector.OnRequest(func(request *colly.Request){
		request.Headers.Set("referer","https://fund.eastmoney.com/")
	})

	var total int
	collector.OnResponse(func(response *colly.Response) {
		var values RespData
		err := json.Unmarshal(response.Body, &values)
		if err != nil {
			println(err.Error())
		}
		total = values.TotalCount
	})
	_ = collector.Visit(fmt.Sprintf(URL, code, 1))
	collector.OnResponse(func (response *colly.Response){
		var values RespData
		err := json.Unmarshal(response.Body, &values)
		if err != nil {
			println(err.Error())
		}
		list := values.Data.LSJZList
		var yv []float64
		var xv []time.Time
		for i := range  list{
			dwjz := list[i].DWJZ
			s := dwjz.(string)
			float, _ := strconv.ParseFloat(s, 2)
			yv = append(yv, float)
			xv = append(xv,time.Time(list[i].FSRQ))
		}
		fund := chart.TimeSeries{
			Style: chart.Style{
				StrokeColor: chart.GetDefaultColor(0),
			},
			XValues: xv,
			YValues: yv,
		}
		font := getFont("./font.ttf")
		graph := chart.Chart{
			XAxis: chart.XAxis{
				Name: "时间",
				NameStyle: chart.Style{
					FontSize: 12,
					Font: font,
				},
				ValueFormatter: chart.TimeDateValueFormatter,
				GridMajorStyle: chart.Style{
					StrokeColor: chart.ColorAlternateGray,
					StrokeWidth: 1.0,
				},
			},
			YAxis: chart.YAxis{
				Name: "基金净值",
				NameStyle: chart.Style{
					FontSize: 12,
					Font: font,
				},
				ValueFormatter: func(v interface{}) string {
					return fmt.Sprintf("%.2f", v.(float64))
				},
			},
			Series: []chart.Series{
				fund,
			},
		}

		f, err := os.Create(fmt.Sprintf("./%s.png",code))
		if err != nil{
			print(err.Error())
		}
		defer f.Close()
		err = graph.Render(chart.PNG, f)
		if err != nil{
			print(err.Error())
		}
	})
	_ = collector.Visit(fmt.Sprintf(URL, code, total))
}


