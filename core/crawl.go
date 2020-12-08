package core

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"io/ioutil"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

const PageUnit = 100
const CrawlRatePerSecond = 10

func requestSearchPage(pageIndex int) string {
	apiEndpoint := "https://work.mma.go.kr/caisBYIS/search/byjjecgeomsaek.do"
	payload := url.Values{}
	payload.Set("al_eopjong_gbcd", "11111,11112")
	payload.Set("eopjong_gbcd_list", "11111,11112")
	payload.Set("eopjong_gbcd", "1")
	payload.Set("eopjong_cd", "11111")
	payload.Set("eopjong_cd", "11112")
	payload.Set("gegyumo_cd", "")
	payload.Set("eopche_nm", "")
	payload.Set("sigungu_addr", "")
	payload.Set("sido_addr", "")
	payload.Set("sigungu_addr", "")
	payload.Set("chaeyongym", "")
	payload.Set("bjinwonym", "")
	payload.Set("searchCondition", "")
	payload.Set("searchKeyword", "")
	payload.Set("pageUnit", strconv.Itoa(PageUnit))
	payload.Set("pageIndex", strconv.Itoa(pageIndex))
	payload.Set("menu_id", "")

	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodPost, apiEndpoint, strings.NewReader(payload.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(payload.Encode())))
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")


	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
}

func requestDetailPage(id int) (string, error) {
	pageUrl := fmt.Sprintf("https://work.mma.go.kr/caisBYIS/search/byjjecgeomsaekView.do?byjjeopche_cd=%d", id)

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, pageUrl, nil)

	resp, _ := client.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body), nil
}

func GetLengthInfo() (pageLength int, corporationLength int) {
	htmlDoc, _ := htmlquery.Parse(strings.NewReader(requestSearchPage(1)))

	resultNode, _ := htmlquery.QueryAll(htmlDoc, "//*[@id=\"content\"]/div[1]/div")
	resultString := resultNode[0].FirstChild.Data
	resultString = regexp.MustCompile("총 게시물 : ([0-9,]+?)건").FindStringSubmatch(resultString)[1]
	resultString = strings.Replace(resultString, ",", "", -1)
	resultInt, _ := strconv.Atoi(resultString)
	return int(math.Ceil(float64(resultInt) / float64(PageUnit))), resultInt
}

func GetSingleCorporationList(pageIndex int) []int {
	page := requestSearchPage(pageIndex)

	page = strings.ReplaceAll(page, "\r", "")
	page = strings.ReplaceAll(page, "\n", "")
	htmlDoc, _ := htmlquery.Parse(strings.NewReader(page))
	corporations, _ := htmlquery.QueryAll(htmlDoc, "//*[@id=\"content\"]/table/tbody/tr/th")

	corporationIdList := make([]int, 0, len(corporations))

	for _, node := range corporations {
		_corporationHref := node.FirstChild.Attr[0].Val
		corporationId, _ := strconv.Atoi(regexp.MustCompile("&byjjeopche_cd=(\\d+?)&").FindStringSubmatch(_corporationHref)[1])
		corporationIdList = append(corporationIdList, corporationId)
	}
	return corporationIdList
}

func CompleteElementWithDetailPage(corporationId int) *AgentableCorporation {
	page, err := requestDetailPage(corporationId)

	if err != nil {
		return nil
	}

	htmlDoc, _ := htmlquery.Parse(strings.NewReader(page))
	regexPeopleCount := regexp.MustCompile("(\\d+?)명")

	parseSimpleNode := func(tableIdx, trIdx ,tdIdx int) string {
		defer func() {
			if v := recover(); v != nil {
				return
			}
		}()
		nodePath := fmt.Sprintf("//*[@id=\"content\"]/div[%d]/table/tbody/tr[%d]/td[%d]", tableIdx, trIdx, tdIdx)
		return htmlquery.Find(htmlDoc, nodePath)[0].FirstChild.Data
	}

	refinePeopleCountString := func(str string) int {
		defer func() {
			if v := recover(); v != nil {
				return
			}
		}()
		intStr := regexPeopleCount.FindStringSubmatch(str)[1]
		result, _ := strconv.Atoi(intStr)
		return result
	}

	element := AgentableCorporation{}

	element.Id = corporationId
	element.Name = parseSimpleNode(1, 1, 1)
	element.BusinessType = parseSimpleNode(2,1, 1)
	element.CorporationSize = parseSimpleNode(2,2, 1)
	element.HyunyukBaejung = refinePeopleCountString(parseSimpleNode(2,3, 1))
	element.HyunyukPyunip = refinePeopleCountString(parseSimpleNode(2,4, 1))
	element.HyunukBokmu = refinePeopleCountString(parseSimpleNode(2,5, 1))
	element.BochungyukBaejung = refinePeopleCountString(parseSimpleNode(2,3, 2))
	element.BochungyukPyunip = refinePeopleCountString(parseSimpleNode(2,4, 2))
	element.BochungyukBokmu = refinePeopleCountString(parseSimpleNode(2,5, 2))

	return &element
}

func CrawlAllCorporationList(totalCorporationLength, pageLength int) [] int {
	corporationList := make([]int, 0, totalCorporationLength)
	corporationTempList := make(chan []int, pageLength)
	var wg sync.WaitGroup

	wg.Add(pageLength)

	for i := 1; i <= pageLength; i++ {
		go func(page int, corpListChannel chan []int) {
			corpListChannel <- GetSingleCorporationList(page)
			wg.Done()
		}(i, corporationTempList)
	}

	go func() {
		wg.Wait()
		close(corporationTempList)
	}()

	for corp := range corporationTempList{
		corporationList = append(corporationList, corp...)
	}

	return corporationList
}

func CrawlAllCorporations() []AgentableCorporation {
	pageLength, corporationLength := GetLengthInfo()
	corporationIdList := CrawlAllCorporationList(corporationLength, pageLength)

	corporationTempList := make(chan AgentableCorporation, 30)
	corporationList := make([]AgentableCorporation, 0, corporationLength)

	var wg sync.WaitGroup

	wg.Add(len(corporationIdList))

	// get detail
	for _, corporationId := range corporationIdList {
		go func() {
			if corporation := CompleteElementWithDetailPage(corporationId); corporation != nil {
				corporationTempList <- *corporation
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(corporationTempList)
	}()

	// detail channel to list
	for corporation := range corporationTempList {
		corporationList = append(corporationList, corporation)
	}


	return corporationList
}
