package search

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"

	"seeder/src/qbittorrent/pkg"
	"seeder/src/qbittorrent/pkg/model"
)

type Client struct {
	BaseUrl string
	Client  *http.Client
	Logger  logrus.FieldLogger
}

func (c Client) Start(pattern string, plugins, categories []string) (int, error) {
	params := url.Values{}
	params.Add("pattern", pattern)
	params.Add("plugins", strings.Join(plugins, "|"))
	params.Add("category", strings.Join(categories, "|"))
	var res struct {
		ID int `json:"id"`
	}
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/start?"+params.Encode(), nil); err != nil {
		return 0, err
	}
	return res.ID, nil
}

func (c Client) Stop(id int) error {
	params := url.Values{}
	params.Add("id", strconv.Itoa(id))
	if err := pkg.Post(c.Client, c.BaseUrl+"/stop?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) GetStatus(id int) (*model.SearchStatus, error) {
	params := url.Values{}
	params.Add("id", strconv.Itoa(id))
	var res []*model.SearchStatus
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/status?"+params.Encode(), nil); err != nil {
		return nil, err
	}
	if len(res) < 1 {
		return nil, fmt.Errorf("response did not contain any statuses")
	}
	return res[0], nil
}

func (c Client) GetStatuses() ([]*model.SearchStatus, error) {
	var res []*model.SearchStatus
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/status", nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (c Client) GetResults(id, limit, offset int) (*model.SearchResultsPaging, error) {
	params := url.Values{}
	params.Add("id", strconv.Itoa(id))
	params.Add("limit", strconv.Itoa(limit))
	params.Add("offset", strconv.Itoa(offset))
	var res model.SearchResultsPaging
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/results?"+params.Encode(), nil); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c Client) Delete(id int) error {
	params := url.Values{}
	params.Add("id", strconv.Itoa(id))
	if err := pkg.Post(c.Client, c.BaseUrl+"/delete?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) GetCategories(plugins []string) ([]string, error) {
	endpoint := c.BaseUrl + "/categories"
	if plugins != nil {
		params := url.Values{}
		params.Add("pluginName", strings.Join(plugins, "|"))
		endpoint += "?" + params.Encode()
	}
	var res []string
	if err := pkg.GetInto(c.Client, &res, endpoint, nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (c Client) GetPlugins() ([]model.SearchPlugin, error) {
	var res []model.SearchPlugin
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/plugins", nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (c Client) InstallPlugins(sources []string) error {
	params := url.Values{}
	params.Add("sources", strings.Join(sources, "|"))
	if err := pkg.Post(c.Client, c.BaseUrl+"/installPlugin?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) UninstallPlugins(plugins []string) error {
	params := url.Values{}
	params.Add("names", strings.Join(plugins, "|"))
	if err := pkg.Post(c.Client, c.BaseUrl+"/uninstallPlugin?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) EnablePlugins(plugins []string, enable bool) error {
	params := url.Values{}
	params.Add("names", strings.Join(plugins, "|"))
	params.Add("enable", fmt.Sprintf("%v", enable))
	if err := pkg.Post(c.Client, c.BaseUrl+"/enablePlugin?"+params.Encode(), nil); err != nil {
		return err
	}
	return nil
}

func (c Client) updatePlugins() error {
	if err := pkg.Post(c.Client, c.BaseUrl+"/updatePlugins", nil); err != nil {
		return err
	}
	return nil
}
