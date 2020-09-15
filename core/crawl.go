package core

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"go.uber.org/ratelimit"
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

func _requestSearchPage(pageIndex int) string {
	apiEndpoint := "https://work.mma.go.kr/caisBYIS/search/byjjecgeomsaek.do"
	payload := url.Values{}
	payload.Set("al_eopjong_gbcd", "11111,11112")
	payload.Set("eopjong_gbcd_list", "11111,11112")
	payload.Set("eopjong_gbcd", "1")
	payload.Set("eopjong_cd", "11111")
	payload.Set("eopjong_cd", "11111")
	payload.Set("gegyumo_cd", "")
	payload.Set("eopche_nm", "")
	payload.Set("juso", "")
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

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
}

func _requestDetailPage(id int) string {
	pageUrl := fmt.Sprintf("https://work.mma.go.kr/caisBYIS/search/byjjecgeomsaekView.do?byjjeopche_cd=%d", id)

	client := &http.Client{}
	req, _ := http.NewRequest(http.MethodGet, pageUrl, nil)

	resp, _ := client.Do(req)

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	return string(body)
}

func getLengthInfo() (pageLength int, corporationLength int) {
	htmlBody := _requestSearchPage(1)
	htmlDoc, _ := htmlquery.Parse(strings.NewReader(htmlBody))

	resultNode, _ := htmlquery.QueryAll(htmlDoc, "//*[@id=\"content\"]/div[1]/div")
	resultString := resultNode[0].FirstChild.Data
	resultString = regexp.MustCompile("총 게시물 : (\\d+?)건").FindStringSubmatch(resultString)[1]
	resultInt, _ := strconv.Atoi(resultString)
	return int(math.Ceil(float64(resultInt) / float64(PageUnit))), resultInt
}

func getSingleCorporationList(pageIndex int) []AgentableCorporation {
	htmlBody := _requestSearchPage(pageIndex)
	corporationList := make([]AgentableCorporation, 0, 0)
	htmlDoc, _ := htmlquery.Parse(strings.NewReader(htmlBody))

	corporations, _ := htmlquery.QueryAll(htmlDoc, "//*[@id=\"content\"]/table/tbody/tr/th")
	for _, node := range corporations {
		corporationName := node.FirstChild.FirstChild.Data
		_corporationHref := node.FirstChild.Attr[0].Val
		corporationId, _ := strconv.Atoi(regexp.MustCompile("&byjjeopche_cd=(\\d+?)&").FindStringSubmatch(_corporationHref)[1])
		corporationList = append(corporationList, AgentableCorporation{
			Name: corporationName,
			Id:   corporationId,
		})
	}
	return corporationList
}

func completeElementWithDetailPage(element *AgentableCorporation, w *sync.WaitGroup) {
	defer w.Done()
	htmlBody := _requestDetailPage(element.Id)
	htmlDoc, _ := htmlquery.Parse(strings.NewReader(htmlBody))
	regexPeopleCount := regexp.MustCompile("(\\d+?)명")

	_parseSimpleNode := func(trIdx int, tdIdx int) string {
		defer func() {
			if v := recover(); v != nil {
				return
			}
		}()
		nodePath := fmt.Sprintf("//*[@id=\"content\"]/div[2]/table/tbody/tr[%d]/td[%d]", trIdx, tdIdx)
		return htmlquery.Find(htmlDoc, nodePath)[0].FirstChild.Data
	}
	_refinePeopleCountString := func(str string) int {
		defer func() {
			if v := recover(); v != nil {
				return
			}
		}()
		intStr := regexPeopleCount.FindStringSubmatch(str)[1]
		result, _ := strconv.Atoi(intStr)
		return result
	}

	element.BusinessType = _parseSimpleNode(1, 1)
	element.CorporationSize = _parseSimpleNode(2, 1)
	element.HyunyukBaejung = _refinePeopleCountString(_parseSimpleNode(3, 1))
	element.HyunyukPyunip = _refinePeopleCountString(_parseSimpleNode(4, 1))
	element.HyunukBokmu = _refinePeopleCountString(_parseSimpleNode(5, 1))
	element.BochungyukBaejung = _refinePeopleCountString(_parseSimpleNode(3, 2))
	element.BochungyukPyunip = _refinePeopleCountString(_parseSimpleNode(4, 2))
	element.BochungyukBokmu = _refinePeopleCountString(_parseSimpleNode(5, 2))
}

func crawlAllCorporations() []AgentableCorporation {
	pageLength, corporationLength := getLengthInfo()
	corporationList := make([]AgentableCorporation, 0)

	for i := 1; i <= pageLength; i++ {
		corporationList = append(corporationList, getSingleCorporationList(i)...)
	}

	wait := new(sync.WaitGroup)
	rl := ratelimit.New(CrawlRatePerSecond)
	wait.Add(corporationLength)
	for i := 0; i < len(corporationList); i++ {
		rl.Take()
		go completeElementWithDetailPage(&corporationList[i], wait)
	}

	wait.Wait()

	return corporationList
}
