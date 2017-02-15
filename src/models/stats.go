package models

import (
	"bytes"
	"encoding/json"

	"github.com/antigloss/go/logger"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/mem"
)

/*GetStats Function called to extract statistics from the database.
 */
func GetStats(packet *MsgClientCmd) ([]byte, error) {

	logger.Trace(packet.Username + " is requesting statistics.")

	// check if the user has access rights

	access, err := UserHasRight([]byte(packet.Username), []byte(packet.Password), "STATS-read")
	if err != nil || access == false {
		logger.Warn("Access denied: User " + packet.Username + " reading STATS")
		return PrepMessageForUser("You do not have access to the Statistics."), err
	}

	logger.Trace("Access granted to statistics.")

	// Grab the current stats.
	DBstats := DB.Bolt.Stats()
	data, _ := json.Marshal(DBstats)

	buf := new(bytes.Buffer)

	/* also grab stats about the server like cpu%, memory usage, disk usage */
	server := serverStats()

	buf.WriteString("{ \"action\":\"stats\", \"server\":" + server + ", \"database\":" + string(data) + " }")

	logger.Trace(" Stats: " + buf.String())

	return buf.Bytes(), nil

}

/*ServerStats this function will gather statistic about the server that running
this application such as CPU, DISK and MEMORY usage and build a JSON object
that hold this information. The return string contain the JSON object that
can be sent to the front end to display the status of the server.
*/
func serverStats() string {

	// memory!
	v, _ := mem.VirtualMemory()
	memoryJSON, err := json.Marshal(*v)
	if err != nil {
		return ""
	}

	cpuInfo, _ := cpu.Info() //([]InfoStat, error)
	cpuInfoJSON, err := json.Marshal(cpuInfo)
	if err != nil {
		return ""
	}

	cpuPercent, _ := cpu.Percent(0, true)
	cpuPercentJSON, err := json.Marshal(cpuPercent)
	if err != nil {
		return ""
	}

	diskInfo, _ := disk.Usage("./")
	diskInfoJSON, err := json.Marshal(*diskInfo)
	if err != nil {
		return ""
	}

	hostInfo, _ := host.Info()
	hostInfoJSON, err := json.Marshal(*hostInfo)
	if err != nil {
		return ""
	}

	return "{\"memory\":" + string(memoryJSON) + ", \"cpuInfo\":" + string(cpuInfoJSON) + ",\"cpuPercent\":" + string(cpuPercentJSON) +
		", \"disk\":" + string(diskInfoJSON) + ", \"host\":" + string(hostInfoJSON) + "}"

}

/*

type InfoStat struct {
    Hostname             string `json:"hostname"`
    Uptime               uint64 `json:"uptime"`
    BootTime             uint64 `json:"bootTime"`
    Procs                uint64 `json:"procs"`           // number of processes
    OS                   string `json:"os"`              // ex: freebsd, linux
    Platform             string `json:"platform"`        // ex: ubuntu, linuxmint
    PlatformFamily       string `json:"platformFamily"`  // ex: debian, rhel
    PlatformVersion      string `json:"platformVersion"` // version of the complete OS
    KernelVersion        string `json:"kernelVersion"`   // version of the OS kernel (if available)
    VirtualizationSystem string `json:"virtualizationSystem"`
    VirtualizationRole   string `json:"virtualizationRole"` // guest or host
    HostID               string `json:"hostid"`             // ex: uuid
}



type VirtualMemoryStat struct {
    // Total amount of RAM on this system
    Total uint64 `json:"total"`

    // RAM available for programs to allocate
    //
    // This value is computed from the kernel specific values.
    Available uint64 `json:"available"`

    // RAM used by programs
    //
    // This value is computed from the kernel specific values.
    Used uint64 `json:"used"`

    // Percentage of RAM used by programs
    //
    // This value is computed from the kernel specific values.
    UsedPercent float64 `json:"usedPercent"`

    // This is the kernel's notion of free memory; RAM chips whose bits nobody
    // cares about the value of right now. For a human consumable number,
    // Available is what you really want.
    Free uint64 `json:"free"`

    // OS X / BSD specific numbers:
    // http://www.macyourself.com/2010/02/17/what-is-free-wired-active-and-inactive-system-memory-ram/
    Active   uint64 `json:"active"`
    Inactive uint64 `json:"inactive"`
    Wired    uint64 `json:"wired"`

    // Linux specific numbers
    // https://www.centos.org/docs/5/html/5.1/Deployment_Guide/s2-proc-meminfo.html
    // https://www.kernel.org/doc/Documentation/filesystems/proc.txt
    Buffers      uint64 `json:"buffers"`
    Cached       uint64 `json:"cached"`
    Writeback    uint64 `json:"writeback"`
    Dirty        uint64 `json:"dirty"`
    WritebackTmp uint64 `json:"writebacktmp"`
}


 // cpu

type InfoStat struct {
    CPU        int32    `json:"cpu"`
    VendorID   string   `json:"vendorId"`
    Family     string   `json:"family"`
    Model      string   `json:"model"`
    Stepping   int32    `json:"stepping"`
    PhysicalID string   `json:"physicalId"`
    CoreID     string   `json:"coreId"`
    Cores      int32    `json:"cores"`
    ModelName  string   `json:"modelName"`
    Mhz        float64  `json:"mhz"`
    CacheSize  int32    `json:"cacheSize"`
    Flags      []string `json:"flags"`
}

cpuPercent array of float64

 // disk


type UsageStat struct {
    Path              string  `json:"path"`
    Fstype            string  `json:"fstype"`
    Total             uint64  `json:"total"`
    Free              uint64  `json:"free"`
    Used              uint64  `json:"used"`
    UsedPercent       float64 `json:"usedPercent"`
    InodesTotal       uint64  `json:"inodesTotal"`
    InodesUsed        uint64  `json:"inodesUsed"`
    InodesFree        uint64  `json:"inodesFree"`
    InodesUsedPercent float64 `json:"inodesUsedPercent"`
}
*/
