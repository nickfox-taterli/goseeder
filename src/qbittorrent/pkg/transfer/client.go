package transfer

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"

	"seeder/src/qbittorrent/pkg"
	"seeder/src/qbittorrent/pkg/model"
)

type Client struct {
	BaseUrl string
	Client  *http.Client
	Logger  logrus.FieldLogger
}

func (c Client) GetTransferInfo() (*model.TransferInfo, error) {
	var res model.TransferInfo
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/info", nil); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c Client) AlternativeSpeedLimitsEnabled() (bool, error) {
	var res int
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/speedLimitsMode", nil); err != nil {
		return false, err
	}
	return res == 1, nil
}

func (c Client) ToggleAlternativeSpeedLimits() error {
	if err := pkg.Post(c.Client, c.BaseUrl+"/toggleSpeedLimitsMode", nil); err != nil {
		return err
	}
	return nil
}

func (c Client) GetGlobalDownloadLimit() (int, error) {
	var res int
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/downloadLimit", nil); err != nil {
		return 0, err
	}
	return res, nil
}

func (c Client) SetGlobalDownloadLimit(limit int) error {
	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	endpoint := c.BaseUrl + "/setDownloadLimit?" + params.Encode()
	if err := pkg.Post(c.Client, endpoint, nil); err != nil {
		return err
	}
	return nil
}

func (c Client) GetGlobalUploadLimit() (int, error) {
	var res int
	if err := pkg.GetInto(c.Client, &res, c.BaseUrl+"/uploadLimit", nil); err != nil {
		return 0, err
	}
	return res, nil
}

func (c Client) SetGlobalUploadLimit(limit int) error {
	params := url.Values{}
	params.Add("limit", strconv.Itoa(limit))
	endpoint := c.BaseUrl + "/setUploadLimit?" + params.Encode()
	if err := pkg.Post(c.Client, endpoint, nil); err != nil {
		return err
	}
	return nil
}
