package qbittorrent

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/publicsuffix"

	"seeder/src/qbittorrent/pkg/application"
	"seeder/src/qbittorrent/pkg/log"
	"seeder/src/qbittorrent/pkg/rss"
	"seeder/src/qbittorrent/pkg/search"
	"seeder/src/qbittorrent/pkg/sync"
	"seeder/src/qbittorrent/pkg/torrent"
	"seeder/src/qbittorrent/pkg/transfer"
)

type Client struct {
	baseURL     string
	logger      logrus.FieldLogger
	client      *http.Client
	Application application.Client
	Log         log.Client
	RSS         rss.Client
	Search      search.Client
	Sync        sync.Client
	Torrent     torrent.Client
	Transfer    transfer.Client
}

func NewClient(baseURL string, logger logrus.FieldLogger) *Client {
	logger = logger.WithField("component", "QBitTorrent Client")
	baseURL = baseURL + "/api/v2"
	client := &http.Client{}
	return &Client{
		baseURL: baseURL,
		logger:  logger,
		client:  client,
		Application: application.Client{
			BaseUrl: baseURL + "/app",
			Client:  client,
			Logger:  logger.WithField("scope", "application"),
		},
		Log: log.Client{
			BaseUrl: baseURL + "/log",
			Client:  client,
			Logger:  logger.WithField("scope", "log"),
		},
		RSS: rss.Client{
			BaseUrl: baseURL + "/rss",
			Client:  client,
			Logger:  logger.WithField("scope", "rss"),
		},
		Search: search.Client{
			BaseUrl: baseURL + "/search",
			Client:  client,
			Logger:  logger.WithField("scope", "search"),
		},
		Sync: sync.Client{
			BaseUrl: baseURL + "/sync",
			Client:  client,
			Logger:  logger.WithField("scope", "sync"),
		},
		Torrent: torrent.Client{
			BaseUrl: baseURL + "/torrents",
			Client:  client,
			Logger:  logger.WithField("scope", "torrents"),
		},
		Transfer: transfer.Client{
			BaseUrl: baseURL + "/transfer",
			Client:  client,
			Logger:  logger.WithField("scope", "transfer"),
		},
	}
}

func (c *Client) Login(username, password string) error {
	endpoint := c.baseURL + "/auth/login"
	data := url.Values{}
	data.Add("username", username)
	data.Add("password", password)
	request, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	request.Header.Add("content-type", "application/x-www-form-urlencoded")
	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			c.logger.Error(err)
		}
	}()
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status %s", resp.Status)
	}
	if len(resp.Cookies()) < 1 {
		return fmt.Errorf("no cookies in login response")
	}
	apiURL, err := url.Parse(c.baseURL)
	if err != nil {
		return err
	}
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return err
	}
	jar.SetCookies(apiURL, []*http.Cookie{resp.Cookies()[0]})
	c.client.Jar = jar
	return nil
}

func (c Client) Logout() error {
	endpoint := c.baseURL + "/auth/logout"
	request, err := http.NewRequest("POST", endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(request)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid status %s", resp.Status)
	}
	return nil
}
