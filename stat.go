package main

import (
	"fmt"
	"github.com/tomcraven/gotable"
	"os"
	"seeder/src/config"
	"seeder/src/qbittorrent"
	"seeder/src/qbittorrent/pkg/model"
	"time"
)

func main() {
	var servers []qbittorrent.Server
	if cfg, err := config.GetConfig(); err == nil {
		for _, value := range cfg.Server {
			if value.Enable == true {
				server := qbittorrent.NewClientWrapper(value.Endpoint, value.Username, value.Password, value.Remark, value.Rule)
				servers = append(servers, server)
			}
		}
	} else {
		os.Exit(1)
	}

	for true {
		t := gotable.NewTable([]gotable.Column{
			gotable.NewColumn("Name", 20),
			gotable.NewColumn("FreeDisk(GB)", 20),
			gotable.NewColumn("DiskLatency(ms)", 20),
			gotable.NewColumn("CurrentSpeed(MB/s)", 20),
			gotable.NewColumn("TaskList(Count)", 20),
			gotable.NewColumn("Transfer(TB)", 20),
			gotable.NewColumn("Ratio(%)", 20),
		})

		for _, server := range servers {
			if r, err := server.Client.Sync.GetMainData(); err == nil {
				ConcurrentDownload := 0
				ConcurrentUpload := 0
				TaskCount := 0

				var options model.GetTorrentListOptions
				options.Filter = "all"
				if ts, err := server.Client.Torrent.GetList(&options); err == nil {
					for _, t := range ts {
						if t.AmountLeft != 0 {
							ConcurrentDownload++
						}
						if t.Upspeed > 0 {
							ConcurrentUpload++
						}
						TaskCount++
						server.Status.EstimatedQuota -= t.AmountLeft
					}
				} else {
					//如果无法获取状态,直接让并行任务数显示最大以跳过规则.
					server.Status.ConcurrentDownload = 65535
				}

				t.Push(
					server.Remark,
					fmt.Sprintf("%.2f", float64(r.ServerState.FreeSpaceOnDisk)/1073741820),
					fmt.Sprintf("%d", r.ServerState.AverageTimeQueue),
					fmt.Sprintf("%.2f(U)|%.2f(D)", float64(r.ServerState.UpInfoSpeed)/1048576.0, float64(r.ServerState.DlInfoSpeed)/1048576.0),
					fmt.Sprintf("%d(U)|%d(D)|%d(A)", ConcurrentUpload, ConcurrentDownload, TaskCount),
					fmt.Sprintf("%.2f(U)|%.2f(D)", float64(r.ServerState.AlltimeUl)/1099511623680, float64(r.ServerState.AlltimeDl)/1099511623680),
					fmt.Sprintf("%.2f", float64(r.ServerState.GlobalRatio)),
				)
			}
		}

		fmt.Printf("\x1bc")
		fmt.Println("QB服务器最新状态:")
		t.Print()

		time.Sleep(5)
	}
}
