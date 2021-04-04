package log

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/google/go-querystring/query"
	"github.com/sirupsen/logrus"

	"seeder/src/qbittorrent/pkg"
	"seeder/src/qbittorrent/pkg/model"
)

type Client struct {
	BaseUrl string
	Client  *http.Client
	Logger  logrus.FieldLogger
}

func (c Client) GetLog(options *model.GetLogOptions) ([]*model.LogEntry, error) {
	endpoint := c.BaseUrl + "/main"
	if options != nil {
		params, err := query.Values(options)
		if err != nil {
			return nil, err
		}
		endpoint += "?" + params.Encode()
	}
	var res []*model.LogEntry
	if err := pkg.GetInto(c.Client, &res, endpoint, nil); err != nil {
		return nil, err
	}
	return res, nil
}

func (c Client) GetPeerLog(lastKnownID int) ([]*model.PeerLogEntry, error) {
	params := url.Values{}
	params.Add("last_known_id", strconv.Itoa(lastKnownID))
	endpoint := c.BaseUrl + "/peers?" + params.Encode()
	var res []*model.PeerLogEntry
	if err := pkg.GetInto(c.Client, &res, endpoint, nil); err != nil {
		return nil, err
	}
	return res, nil
}
