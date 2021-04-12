package qbittorrent

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"seeder/src/config"
	"seeder/src/datebase"
	"seeder/src/qbittorrent/pkg/model"
	"strings"
	"time"
)

type ServerStatus struct {
	FreeSpaceOnDisk    int
	EstimatedQuota     int
	ConcurrentDownload int
	UpInfoSpeed        int
	DownInfoSpeed      int
	DiskLatency        int
}

type Server struct {
	Client *Client
	Rule   config.ServerRule
	Remark string
	Status ServerStatus
}

func (s *Server) ServerClean(cfg config.Config, db datebase.Client) {
	//如果不在清理要求.
	if s.Status.FreeSpaceOnDisk >= s.Rule.DiskThreshold {
		return
	}

	//开始执行删除操作(第一圈,删除无效内容,无论如何都不会跳过的.)
	if s.Status.FreeSpaceOnDisk < s.Rule.DiskThreshold {
		var options model.GetTorrentListOptions
		options.Filter = "all"
		if ts, err := s.Client.Torrent.GetList(&options); err == nil {
			for _, t := range ts {
				for _, n := range cfg.Node {
					if n.Source == t.Category {
						if trackers, err := s.Client.Torrent.GetTrackers(t.Hash); err == nil && (int(time.Now().Unix())-t.AddedOn) > s.Rule.MinAliveTime {
							for _, tracker := range trackers {
								if tracker.Status == model.TrackerStatusNotContacted || tracker.Status == model.TrackerStatusNotWorking {
									s.Client.Torrent.DeleteTorrents([]string{t.Hash}, true)
									fmt.Println("[" + s.Remark + "]清理无效种子." + t.Name)
								}
							}
						}
					}
				}
			}
		}
	}
	s.CalcEstimatedQuota()

	//开始执行删除操作(第二圈,删除其中一个最古老的正在进行的任务.)
	MaxAliveTime := 0
	MaxAliveSeeder := ""
	MaxAliveName := ""
	if s.Status.FreeSpaceOnDisk < s.Rule.DiskThreshold {
		var options model.GetTorrentListOptions
		options.Filter = "all"
		if ts, err := s.Client.Torrent.GetList(&options); err == nil {
			for _, t := range ts {
				for _, n := range cfg.Node {
					if n.Source == t.Category {
						if t.AmountLeft != 0 {
							if (int(time.Now().Unix()) - t.CompletionOn) > s.Rule.MaxAliveTime {
								if MaxAliveTime < int(time.Now().Unix())-t.CompletionOn {
									MaxAliveTime = int(time.Now().Unix()) - t.CompletionOn
									MaxAliveSeeder = t.Hash
									MaxAliveName = t.Name
								}
							}
						}
					}
				}
			}
		}
	}
	if MaxAliveTime != 0 {
		s.Client.Torrent.DeleteTorrents([]string{MaxAliveSeeder}, true)
		fmt.Println("[" + s.Remark + "]删除超时种子." + MaxAliveName)
		return
	}

	//开始执行删除操作(第三圈,删除其中一个最古老的完成的任务.)
	MaxAliveTime = 0
	MaxAliveSeeder = ""
	MaxAliveName = ""
	if s.Status.FreeSpaceOnDisk < s.Rule.DiskThreshold {
		var options model.GetTorrentListOptions
		options.Filter = "all"
		if ts, err := s.Client.Torrent.GetList(&options); err == nil {
			for _, t := range ts {
				for _, n := range cfg.Node {
					if n.Source == t.Category {
						if t.AmountLeft == 0 {
							if (int(time.Now().Unix()) - t.CompletionOn) > s.Rule.MaxAliveTime {
								if MaxAliveTime < int(time.Now().Unix())-t.CompletionOn {
									MaxAliveTime = int(time.Now().Unix()) - t.CompletionOn
									MaxAliveSeeder = t.Hash
									MaxAliveName = t.Name
								}
							}
						}
					}
				}
			}
		}
	}
	if MaxAliveTime != 0 {
		s.Client.Torrent.DeleteTorrents([]string{MaxAliveSeeder}, true)
		fmt.Println("[" + s.Remark + "]删除超时种子." + MaxAliveName)
		return
	}

	fmt.Println("[" + s.Remark + "]无法完成清理.")
}

func (s *Server) ServerRuleTest() bool {
	TestStatus := "测试成功"

	if s.Rule.MaxDiskLatency < s.Status.DiskLatency {
		TestStatus = "测试失败"
	}

	if s.Status.UpInfoSpeed > s.Rule.MaxSpeed {
		TestStatus = "测试失败"
	}

	if s.Status.DownInfoSpeed > s.Rule.MaxSpeed {
		TestStatus = "测试失败"
	}

	if s.Status.ConcurrentDownload > s.Rule.ConcurrentDownload {
		TestStatus = "测试失败"
	}

	fmt.Printf("[%s][%s] 当前磁盘空间余量 %.2f[%.2f]GB,磁盘延迟 %d[%d] ms,上传速度 %.2f[%.2f],下载速度 %.2f[%.2f],同时任务 %d[%d] 个.\n",
		s.Remark, TestStatus,
		float64(s.Status.EstimatedQuota)/1073741824.0, float64(s.Status.FreeSpaceOnDisk)/1073741824,
		s.Status.DiskLatency, s.Rule.MaxDiskLatency,
		float64(s.Status.UpInfoSpeed)/1048576.0, float64(s.Rule.MaxSpeed)/1048576.0,
		float64(s.Status.DownInfoSpeed)/1048576.0, float64(s.Rule.MaxSpeed)/1048576.0,
		s.Status.ConcurrentDownload, s.Rule.ConcurrentDownload,
	)

	if TestStatus == "测试失败" {
		return false
	}

	return true

}

func (s *Server) AddTorrentByURL(URL string, Size int) bool {
	var options_add model.AddTorrentsOptions
	options_add.Savepath = "/downloads/"
	options_add.Category = strings.Split(strings.Split(URL, "//")[1], "/")[0]

	var options_list model.GetTorrentListOptions
	options_list.Filter = "all"
	if ts, err := s.Client.Torrent.GetList(&options_list); err == nil {
		for _, t := range ts {
			if t.Size == Size {
				//有同样大小的种子在一个机,容易产生混乱.
				//@TODO后期可以利用这个特性做辅粽功能
				return false
			}
		}
	}

	if Size < s.Rule.MaxTaskSize && Size > s.Rule.MinTaskSize && s.ServerRuleTest() == true {
		if err := s.Client.Torrent.AddURLs([]string{URL}, &options_add); err == nil {
			return true
		}
	}

	return false
}

func (s *Server) CalcEstimatedQuota() {
	// 这里计算出来的是磁盘正在可以用的空间
	if r, err := s.Client.Sync.GetMainData(); err == nil {
		s.Status.DiskLatency = r.ServerState.AverageTimeQueue
		s.Status.FreeSpaceOnDisk = r.ServerState.FreeSpaceOnDisk
		s.Status.EstimatedQuota = r.ServerState.FreeSpaceOnDisk
		// 这里计算出来的是磁盘预期可以用的空间.(假设种子会全部下载)
		var options model.GetTorrentListOptions
		options.Filter = "all"
		if ts, err := s.Client.Torrent.GetList(&options); err == nil {
			s.Status.ConcurrentDownload = 0
			for _, t := range ts {
				if t.AmountLeft != 0 {
					s.Status.ConcurrentDownload++
				}
				s.Status.EstimatedQuota -= t.AmountLeft
			}
		} else {
			//如果无法获取状态,直接让并行任务数显示最大以跳过规则.
			s.Status.ConcurrentDownload = 65535
		}
	}

	if r, err := s.Client.Transfer.GetTransferInfo(); err == nil {
		s.Status.UpInfoSpeed = r.UpInfoSpeed
		s.Status.DownInfoSpeed = r.DlInfoSpeed
	}
}

func NewClientWrapper(baseURL string, username string, password string, remark string, rule config.ServerRule) Server {
	var logger = logrus.New()
	server := NewClient(baseURL, logger)
	server.Login(username, password)

	return Server{
		Client: server,
		Rule:   rule,
		Remark: remark,
	}
}
