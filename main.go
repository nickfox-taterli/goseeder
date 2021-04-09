package main

import (
	"fmt"
	"github.com/robfig/cron"
	"io/ioutil"
	"net/http"
	"os"
	"seeder/src/config"
	"seeder/src/datebase"
	"seeder/src/nexus"
	"seeder/src/qbittorrent"
	"strconv"
)

func checkin() {

}

func main() {
	var db datebase.Client
	var nodes []nexus.Client
	var servers []qbittorrent.Server

	cron := cron.New()

	cron.AddFunc("@every 30s", func() {
		url := "https://api.honeybadger.io/v1/check_in/vOIMxP"

		client := &http.Client{
		}
		req, err := http.NewRequest("GET", url, nil)

		if err != nil {
			res, err := client.Do(req)
			if err == nil {
				defer res.Body.Close()
				ioutil.ReadAll(res.Body)
			}
		}
	})

	if cfg, err := config.GetConfig(); err == nil {
		db = datebase.NewClient(cfg.Db)
		for _, value := range cfg.Node {
			if value.Enable == true {
				node := nexus.NewClient(value.Source, value.Limit, value.Passkey, value.Rule)
				nodes = append(nodes, node)
			}
		}
		for _, value := range cfg.Server {
			if value.Enable == true {
				server := qbittorrent.NewClientWrapper(value.Endpoint, value.Username, value.Password, value.Remark, value.Rule)

				server.CalcEstimatedQuota()
				server.ServerClean(cfg, db)

				cron.AddFunc("@every 5s", func() { server.CalcEstimatedQuota() })
				cron.AddFunc("@every 1m", func() { server.ServerClean(cfg, db) })
				cron.Start()

				servers = append(servers, server)
			}
		}
	} else {
		os.Exit(1)
	}

	for true {
		var ts []nexus.Torrent
		for _, node := range nodes {
			ts, _ = node.Get()
			for _, t := range ts {
				// 解决重复添加问题
				for _, server := range servers {
					server.CalcEstimatedQuota()
					if db.Get(t.GUID) == false {
						if Size, err := strconv.Atoi(t.Size); err == nil {
							if server.AddTorrentByURL(t.URL, Size) == true {
								fmt.Println(server.Remark + "添加了种子:" + t.Title)
								db.Insert(t.Title, t.GUID, t.URL)
							}
						}
					}
				}
			}
		}
	}
}
