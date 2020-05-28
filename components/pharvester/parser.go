package pharvester

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"

	"git.pnhub.ru/core/libs/log"
	"git.pnhub.ru/core/libs/util"

	"golang.org/x/net/html"
)

var IpRegexp = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
var PortRegexp = regexp.MustCompile(`\b([1-9]|[1-5]?[0-9]{2,4}|6[1-4][0-9]{3}|65[1-4][0-9]{2}|655[1-2][0-9]|6553[1-5])\b`)
var AddressRegexp = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b:\b([1-9]|[1-5]?[0-9]{2,4}|6[1-4][0-9]{3}|65[1-4][0-9]{2}|655[1-2][0-9]|6553[1-5])\b`)
var ProtocolRegexp = regexp.MustCompile(`(?i)(socks4|socks5|https|http)`)
var CleanRegexp = regexp.MustCompile(`[\t\r\n]+`)

const Host = "Host"
const Port = "Port"
const Protocol = "Protocol"
const Address = "Address"

type NodeVisitCallback = func(n *html.Node) bool

type Parser struct {
	ctx    context.Context
	logger log.Logger
}

func NewParser(ctx context.Context, logger log.Logger) *Parser {
	return &Parser{ctx: ctx, logger: logger}
}

func (p *Parser) Parse(resp *colly.Response, cfg *SourceConfig) ([]*Proxy, error) {
	var buf = bytes.NewBuffer(resp.Body)
	d, err := goquery.NewDocumentFromReader(buf)
	if err != nil {
		return nil, err
	}
	u := resp.Request.URL.String()
	return p.ExtractData(d, u, cfg)
}

func (p *Parser) ExtractList(d *goquery.Document, selector *Selector) []string {
	var filter *regexp.Regexp
	if selector.FilterRegexp != "" {
		filter = regexp.MustCompile(selector.FilterRegexp)
	} else {
		switch selector.Target {
		case Host:
			filter = IpRegexp
		case Port:
			filter = PortRegexp
		case Protocol:
			filter = ProtocolRegexp
		case Address:
			filter = AddressRegexp
		default:
			filter = regexp.MustCompile(".*")
		}
	}
	var mappers = make(map[*regexp.Regexp]string)
	for k, v := range selector.Mapping {
		matcher := regexp.MustCompile(k)
		mappers[matcher] = v
	}

	var dataList = make([]string, 0)
	s := d.Find(selector.Selector)
	s.Each(func(i int, selection *goquery.Selection) {
		text := selection.Text()
		text = CleanRegexp.ReplaceAllString(strings.TrimSpace(text), "")
		for k, v := range mappers {
			if k.Match([]byte(text)) {
				text = k.ReplaceAllString(text, v)
			}
		}
		p.logger.Debugf("[target: %s | selector: '%s'] -> %s", selector.Target, selector.Selector, text)
		text = filter.FindString(text)
		if !filter.Match([]byte(text)) {
			return
		}
		dataList = append(dataList, text)
	})
	if selector.EnableValidation {
		if len(dataList) == 0 {
			panic(fmt.Sprintf("there is no data finded by selector %s", selector.Selector))
		}
		for _, text := range dataList {
			if !filter.Match([]byte(text)) {
				panic(fmt.Sprintf("data:[%s] do not match %s", text, filter.String()))
			}
		}
	}
	return dataList
}

func (p *Parser) ExtractData(d *goquery.Document, u string, cfg *SourceConfig) ([]*Proxy, error) {
	var listMap = make(map[string][]string)
	for _, selector := range cfg.Selectors {
		p.logger.Infof("URL:%s Target:%s Selector:%s extracting...", u, selector.Target, selector.Selector)
		list := p.ExtractList(d, selector)
		p.logger.Infof("URL:%s Target:%s Selector:%s extracted %d elements", u, selector.Target, selector.Selector, len(list))
		if selector.Target == Address {
			hosts := make([]string, 0)
			ports := make([]string, 0)
			for _, addr := range list {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				hosts = append(hosts, host)
				ports = append(ports, port)
			}
			listMap[Host] = hosts
			listMap[Port] = ports
		} else {
			listMap[selector.Target] = list
		}

	}
	ipList := listMap[Host]
	length := len(ipList)
	for key, list := range listMap {
		diff := length - len(list)
		if diff > 0 {
			list = append(list, make([]string, diff)...)
		}
		listMap[key] = list
	}

	var pList = make([]*Proxy, 0)
	var i = 0
	for {
		var proxy = &Proxy{
			Source:    cfg.URL,
			CreatedAt: time.Now(),
		}

		if i == length {
			return pList, nil
		}
		for key, list := range listMap {
			val := list[i]
			util.MustSetUnexportedFieldByName(key, proxy, val)
		}
		pList = append(pList, proxy)
		i++
	}
}

func (p *Parser) ParseHTML(r io.Reader) (*html.Node, error) {
	return html.Parse(r)
}

func (p *Parser) WalkHTML(node *html.Node, callback NodeVisitCallback) {
	var f func(*html.Node)
	f = func(n *html.Node) {
		ok := callback(n)
		if !ok {
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(node)
}
