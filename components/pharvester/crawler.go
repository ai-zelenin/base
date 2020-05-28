package pharvester

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"

	"github.com/gocolly/colly"

	"git.pnhub.ru/core/libs/log"
)

type Crawler struct {
	ctx     context.Context
	logger  log.Logger
	parser  *Parser
	storage *Storage
}

func NewCrawler(ctx context.Context, logger log.Logger, parser *Parser, storage *Storage) *Crawler {
	return &Crawler{
		ctx:     ctx,
		logger:  log.ForkLogger(logger),
		parser:  parser,
		storage: storage,
	}
}

func (c *Crawler) CrawlSourceURL(cfg *SourceConfig) error {
	followRegexp, err := regexp.Compile(cfg.FollowRegexp)
	if err != nil {
		return err
	}
	u, err := url.Parse(cfg.URL)
	if err != nil {
		return err
	}

	var domain = u.Host
	collector := colly.NewCollector(
		colly.AllowedDomains(domain),
		colly.MaxDepth(cfg.Depth),
		colly.CacheDir(u.Host),
		colly.Async(true),
	)

	err = collector.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Delay:       cfg.Delay,
		RandomDelay: cfg.RandomDelay,
		Parallelism: cfg.Threads,
	})
	if err != nil {
		return err
	}
	// Find and visit all links
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if followRegexp.Match([]byte(link)) {
			err := e.Request.Visit(link)
			if err != nil && err != colly.ErrAlreadyVisited {
				c.logger.Debugf("%s - %v ", link, err)
			}
		}
	})

	collector.OnError(func(response *colly.Response, e error) {
		c.logger.Warn(err)
	})

	collector.OnRequest(func(r *colly.Request) {
		c.logger.Debugf("Visiting %s", r.URL)
	})

	collector.OnScraped(func(response *colly.Response) {
		pl, err := c.parser.Parse(response, cfg)
		if err != nil {
			c.logger.Error(err)
		} else {
			for _, proxy := range pl {
				err = c.storage.Save(proxy)
				if err != nil {
					c.logger.Warn(err)
				}
			}
		}
	})

	err = collector.Visit(cfg.URL)
	if err != nil {
		return err
	}
	collector.Wait()

	return nil
}
func (c *Crawler) CrawlSourceDir(cfg *SourceConfig) error {
	dir, err := filepath.Abs(filepath.Dir(cfg.Dir))
	if err != nil {
		return err
	}

	t := new(http.Transport)
	t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))

	collector := colly.NewCollector()
	collector.WithTransport(t)

	collector.OnRequest(func(r *colly.Request) {
		c.logger.Debugf("Visiting %s", r.URL)
	})

	collector.OnScraped(func(response *colly.Response) {
		pl, err := c.parser.Parse(response, cfg)
		if err != nil {
			c.logger.Error(err)
		} else {
			for _, proxy := range pl {
				err = c.storage.Save(proxy)
				if err != nil {
					c.logger.Warn(err)
				}
			}
		}
	})
	files, err := filepath.Glob(fmt.Sprintf("%s/*", dir))
	if err != nil {
		return err
	}
	for _, file := range files {
		err = collector.Visit("file://" + file)
		if err != nil {
			return err
		}
	}
	collector.Wait()
	return nil
}
