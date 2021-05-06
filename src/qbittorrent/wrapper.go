package qbittorrent

import (
	"fmt"
	"seeder/src/config"
	"seeder/src/datebase"
	"seeder/src/qbittorrent/pkg/model"
	"strconv"
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
	Rule   config.RawServerRule
	Remark string
	Status ServerStatus
}

func (s *Server) ServerClean(cfg config.Config, db datebase.Client) {
	if s.Status.FreeSpaceOnDisk < s.Rule.DiskThreshold {
		if ts, err := s.Client.GetList(); err == nil {
			for _, t := range ts {
				for _, n := range cfg.Node {
					if n.Source == t.Category {
						if trackers, err := s.Client.GetTrackers(t.Hash); err == nil && (int(time.Now().Unix())-t.AddedOn) > s.Rule.MinAliveTime {
							for _, tracker := range trackers {
								if tracker.Status == model.TrackerStatusNotContacted || tracker.Status == model.TrackerStatusNotWorking {
									s.Client.DeleteTorrents(t.Hash)
									fmt.Println("[" + s.Remark + "]清理无效种子." + t.Name)
								}
							}
						}
					}
				}
			}
		}

		//开始执行删除操作(第二圈,删除其中一个最古老的完成的任务.)
		MaxAliveTime := 0
		MaxAliveSeeder := ""
		MaxAliveName := ""

		if ts, err := s.Client.GetList(); err == nil {
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
		if MaxAliveTime != 0 {
			s.Client.DeleteTorrents(MaxAliveSeeder)
			fmt.Println("[" + s.Remark + "]删除超时种子." + MaxAliveName)
			return
		}

		//开始执行删除操作(第三圈,删除其中一个最古老的正在进行的任务.)
		MaxAliveTime = 0
		MaxAliveSeeder = ""
		MaxAliveName = ""
		if ts, err := s.Client.GetList(); err == nil {
			for _, t := range ts {
				for _, n := range cfg.Node {
					if n.Source == t.Category {
						if t.AmountLeft != 0 {
							if (int(time.Now().Unix()) - t.AddedOn) > s.Rule.MaxAliveTime {
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

		if MaxAliveTime != 0 {
			s.Client.DeleteTorrents(MaxAliveSeeder)
			fmt.Println("[" + s.Remark + "]删除超时种子." + MaxAliveName)
			return
		}
	}

	//fmt.Println("[" + s.Remark + "]无法完成清理.")
}

func (s *Server) ServerRuleTest() bool {
	TestStatus := "测试成功"

	if s.Rule.MaxDiskLatency <= s.Status.DiskLatency {
		TestStatus = "测试失败"
	}

	if s.Status.UpInfoSpeed >= s.Rule.MaxSpeed {
		TestStatus = "测试失败"
	}

	if s.Status.DownInfoSpeed >= s.Rule.MaxSpeed {
		TestStatus = "测试失败"
	}

	if s.Status.ConcurrentDownload >= s.Rule.ConcurrentDownload {
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

func (s *Server) AddTorrentByURL(URL string, Size int, SpeedLimit int) bool {
	var options_add model.AddTorrentsOptions
	options_add.Savepath = "/downloads/"
	options_add.Category = strings.Split(strings.Split(URL, "//")[1], "/")[0]
	options_add.DlLimit = strconv.Itoa(SpeedLimit)
	options_add.UpLimit = strconv.Itoa(SpeedLimit)

	if ts, err := s.Client.GetList(); err == nil {
		for _, t := range ts {
			if t.Size == Size {
				//有同样大小的种子在一个机,容易产生混乱.
				//@TODO后期可以利用这个特性做辅粽功能,必须数据库有Size才可以.
				return false
			}
		}
	}

	//测试特殊网站规则(Beta),避免有些网站付费种太多.

	// HDTIME网站不足100G大小种子会被忽略.
	if strings.Contains(URL, "hdtime.org") {
		if Size < 100 * 1024 * 1024 * 1024 {
			return true
		}
	}

	if Size < s.Rule.MaxTaskSize && Size > s.Rule.MinTaskSize && s.ServerRuleTest() == true {
		//如果允许超量提交(即塞了这个任务后,并且任务完成后空间会负数,则不检查空间直接OK!),否则检查是否塞进去后还有空间剩余.
		//这个功能针对极小盘有很好的作用,因为极小盘很容易就会塞满,参数又不好调整.
		if s.Rule.DiskOverCommit == true || (s.Status.EstimatedQuota-Size) > (s.Rule.DiskThreshold/10) {
			if err := s.Client.AddURLs(URL, &options_add); err == nil {
				return true
			}
		}
	}
	return false
}

func (s *Server) CalcEstimatedQuota() {
	// 这里计算出来的是磁盘正在可以用的空间
	if r, err := s.Client.GetMainData(); err == nil {
		s.Status.DiskLatency = r.ServerState.AverageTimeQueue
		s.Status.FreeSpaceOnDisk = r.ServerState.FreeSpaceOnDisk
		s.Status.EstimatedQuota = r.ServerState.FreeSpaceOnDisk
		// 这里计算出来的是磁盘预期可以用的空间.(假设种子会全部下载)
		if ts, err := s.Client.GetList(); err == nil {
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

	if r, err := s.Client.GetTransferInfo(); err == nil {
		s.Status.UpInfoSpeed = r.UpInfoSpeed
		s.Status.DownInfoSpeed = r.DlInfoSpeed
	}
}

func NewClientWrapper(baseURL string, username string, password string, remark string, rule config.ServerRule) Server {
	server,err := NewClient(baseURL,username,password)

	if err != nil {
		print("[" + remark + "]密码打错了,赶紧去修正.")
	}

	return Server{
		Client: server,
		Rule: config.RawServerRule{
			ConcurrentDownload: rule.ConcurrentDownload,
			DiskThreshold:      int(rule.DiskThreshold * 1024 * 1024 * 1024),
			DiskOverCommit:     rule.DiskOverCommit,
			MaxSpeed:           int(rule.MaxSpeed * 1024 * 1024),
			MinAliveTime:       rule.MinAliveTime,
			MaxAliveTime:       rule.MaxAliveTime,
			MinTaskSize:        int(rule.MinTaskSize * 1024 * 1024 * 1024),
			MaxTaskSize:        int(rule.MaxTaskSize * 1024 * 1024 * 1024),
			MaxDiskLatency:     rule.MaxDiskLatency,
		},
		Remark: remark,
	}
}
