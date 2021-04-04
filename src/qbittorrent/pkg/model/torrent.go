package model

type Torrent struct {
	// Completion On Timestamp
	CompletionOn int `json:"completion_on"`
	// Added On Timestamp
	AddedOn int `json:"added_on"`
	// Amount left
	AmountLeft int `json:"amount_left"`
	// Torrent hash
	Hash string `json:"hash"`
	// Torrent name
	Name string `json:"name"`
	// Total size (bytes) of files selected for download
	Size int `json:"size"`
	// Torrent progress (percentage/100)
	Progress float64 `json:"progress"`
	// Torrent download speed (bytes/s)
	Dlspeed int `json:"dlspeed"`
	// Torrent upload speed (bytes/s)
	Upspeed int `json:"upspeed"`
	// Torrent priority. Returns -1 if queuing is disabled or torrent is in seed mode
	Priority int `json:"priority"`
	// Number of seeds connected to
	NumSeeds int `json:"num_seeds"`
	// Number of seeds in the swarm
	NumComplete int `json:"num_complete"`
	// Number of leechers connected to
	NumLeechs int `json:"num_leechs"`
	// Number of leechers in the swarm
	NumIncomplete int `json:"num_incomplete"`
	// Torrent share ratio. Max ratio value: 9999.
	Ratio float64 `json:"ratio"`
	// Torrent ETA (seconds)
	Eta int `json:"eta"`
	// Torrent state. See table here below for the possible values
	State TorrentState `json:"state"`
	// True if sequential download is enabled
	SeqDl bool `json:"seq_dl"`
	// True if first last piece are prioritized
	FLPiecePrio bool `json:"f_l_piece_prio"`
	// Category of the torrent
	Category string `json:"category"`
	// True if super seeding is enabled
	SuperSeeding bool `json:"super_seeding"`
	// True if force start is enabled for this torrent
	ForceStart bool `json:"force_start"`
}

type TorrentState string

const (
	// Some error occurred, applies to paused torrents
	StateError TorrentState = "error"
	// Torrent data files is missing
	StateMissingFiles TorrentState = "missingFiles"
	// Torrent is being seeded and data is being transferred
	StateUploading TorrentState = "uploading"
	// Torrent is paused and has finished downloading
	StatePausedUP TorrentState = "pausedUP"
	// Queuing is enabled and torrent is queued for upload
	StateQueuedUP TorrentState = "queuedUP"
	// Torrent is being seeded, but no connection were made
	StateStalledUP TorrentState = "stalledUP"
	// Torrent has finished downloading and is being checked
	StateCheckingUP TorrentState = "checkingUP"
	// Torrent is forced to uploading and ignore queue limit
	StateForcedUP TorrentState = "forcedUP"
	// Torrent is allocating disk space for download
	StateAllocating TorrentState = "allocating"
	// Torrent is being downloaded and data is being transferred
	StateDownloading TorrentState = "downloading"
	// Torrent has just started downloading and is fetching metadata
	StateMetaDL TorrentState = "metaDL"
	// Torrent is paused and has NOT finished downloading
	StatePausedDL TorrentState = "pausedDL"
	// Queuing is enabled and torrent is queued for download
	StateQueuedDL TorrentState = "queuedDL"
	// Torrent is being downloaded, but no connection were made
	StateStalledDL TorrentState = "stalledDL"
	// Same as checkingUP, but torrent has NOT finished downloading
	StateCheckingDL TorrentState = "checkingDL"
	// Torrent is forced to downloading to ignore queue limit
	StateForceDL TorrentState = "forceDL"
	// Checking resume data on qBt startup
	StateCheckingResumeData TorrentState = "checkingResumeData"
	// Torrent is moving to another location
	StateMoving TorrentState = "moving"
	// Unknown status
	StateUnknown TorrentState = "unknown"
)
