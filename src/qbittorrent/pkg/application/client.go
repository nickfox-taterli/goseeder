package application

import (
	"net/http"

	"github.com/sirupsen/logrus"

	"seeder/src/qbittorrent/pkg"
	"seeder/src/qbittorrent/pkg/model"
)

type Client struct {
	BaseUrl string
	Client  *http.Client
	Logger  logrus.FieldLogger
}

func (c Client) GetAppVersion() (string, error) {
	var res string
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/version", nil); err != nil {
		return "", err
	}
	return res, nil
}

func (c Client) GetAPIVersion() (string, error) {
	var res string
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/webapiVersion", nil); err != nil {
		return "", err
	}
	return res, nil
}

func (c Client) GetBuildInfo() (*model.BuildInfo, error) {
	var res model.BuildInfo
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/buildInfo", nil); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c Client) GetAppPreferences() (*model.Preferences, error) {
	var res model.Preferences
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/preferences", nil); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c Client) SetAppPreferences(p *model.Preferences) error {
	if err := pkg.Post(c.Client, c.BaseUrl+"/setPreferences", p); err != nil {
		return err
	}
	return nil
}

func (c Client) GetDefaultSavePath() (string, error) {
	var res string
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/defaultSavePath", nil); err != nil {
		return "", err
	}
	return res, nil
}
