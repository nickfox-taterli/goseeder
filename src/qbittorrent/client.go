package qbittorrent

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/publicsuffix"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"seeder/src/qbittorrent/pkg/model"
	"strings"

)

type Client struct {
	baseURL     string
	loginURI     string
	client      *http.Client
}


func Auth(baseURL string, loginURI string) (*http.Client,error) {
	client := &http.Client {}
	req, err := http.NewRequest("GET", loginURI, nil)

	if err != nil {
		fmt.Println(err)
		return nil,err
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return nil,err
	}

	if string(body) != "Ok." {
		return nil,errors.New("Password Error!")
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return nil,err
	}
	apiURL, err := url.Parse(baseURL)
	jar.SetCookies(apiURL, []*http.Cookie{res.Cookies()[0]})
	client.Jar = jar

	return client,err
}

func (c *Client) GetInto(url string, target interface{}) (err error) {
	req, err := http.NewRequest("GET", c.baseURL + url, nil)

	if err != nil {
		Auth(c.baseURL,c.loginURI)
		return err
	}

	res, err := c.client.Do(req)
	if err != nil {
		Auth(c.baseURL,c.loginURI)
		return err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		Auth(c.baseURL,c.loginURI)
		return err
	}

	if err := json.NewDecoder(bytes.NewReader(body)).Decode(target); err != nil {
		if err2 := json.NewDecoder(strings.NewReader(`"` + string(body) + `"`)).Decode(target); err2 != nil {
			Auth(c.baseURL,c.loginURI)
			return err
		}
	}

	return nil
}

func (c Client) GetMainData() (*model.SyncMainData, error) {
	var res model.SyncMainData

	err := c.GetInto("/sync/maindata",&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}


func (c Client) GetList() ([]*model.Torrent, error) {
	var res []*model.Torrent
	err := c.GetInto("/torrents/info?filter=all",&res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (c Client) GetTransferInfo() (*model.TransferInfo, error) {
	var res model.TransferInfo
	if err := c.GetInto("/transfer/info",&res) ;err != nil {
		return nil, err
	}
	return &res, nil
}

func (c Client) GetTrackers(hash string) ([]*model.TorrentTracker, error) {
	var res []*model.TorrentTracker
	if err := c.GetInto("/torrents/trackers?hash=" + hash,&res); err != nil {
		return nil, err
	}
	return res, nil
}

func (c Client) DeleteTorrents(hash string) error {
	var res string
	if err := c.GetInto("/torrents/delete?hashes="  + hash + "&deleteFiles=true",&res); err != nil {
		return err
	}
	return nil
}

func (c Client) AddURLs(DestLink string,options *model.AddTorrentsOptions) error {
	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	_ = writer.WriteField("urls", DestLink)
	_ = writer.WriteField("category", options.Category)
	_ = writer.WriteField("savepath", options.Savepath)
	_ = writer.WriteField("upLimit", options.UpLimit)
	_ = writer.WriteField("dlLimit", options.DlLimit)
	err := writer.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	req, err := http.NewRequest("POST", c.baseURL + "/torrents/add", payload)

	if err != nil {
		fmt.Println(err)
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := c.client.Do(req)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if string(body) != "Ok." {
		return errors.New("AddURL Error!")
	}

	return nil
}

func NewClient(baseURL string,username string,password string) (*Client,error) {
	baseURL = baseURL + "/api/v2"
	loginURI := baseURL + "/auth/login?username=" + username + "&password=" + password
	client, err := Auth(baseURL,loginURI)

	c := Client{
		baseURL:  baseURL,
		loginURI: loginURI,
		client:   client,
	}


	return &c, err
}
