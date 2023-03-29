package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/tv42/httpunix"
)

type dockerCreateObject struct {
	Hostname     string        `json:"Hostname"`
	Domainname   string        `json:"Domainname"`
	User         string        `json:"User"`
	AttachStdin  bool          `json:"AttachStdin"`
	AttachStdout bool          `json:"AttachStdout"`
	AttachStderr bool          `json:"AttachStderr"`
	Tty          bool          `json:"Tty"`
	OpenStdin    bool          `json:"OpenStdin"`
	StdinOnce    bool          `json:"StdinOnce"`
	Env          []interface{} `json:"Env"`
	Cmd          []string      `json:"Cmd"`
	Image        string        `json:"Image"`
	Volumes      struct {
	} `json:"Volumes"`
	WorkingDir string      `json:"WorkingDir"`
	Entrypoint interface{} `json:"Entrypoint"`
	OnBuild    interface{} `json:"OnBuild"`
	Labels     struct {
	} `json:"Labels"`
	HostConfig struct {
		Binds           interface{} `json:"Binds"`
		ContainerIDFile string      `json:"ContainerIDFile"`
		LogConfig       struct {
			Type   string `json:"Type"`
			Config struct {
			} `json:"Config"`
		} `json:"LogConfig"`
		NetworkMode  string `json:"NetworkMode"`
		PortBindings struct {
		} `json:"PortBindings"`
		RestartPolicy struct {
			Name              string `json:"Name"`
			MaximumRetryCount int    `json:"MaximumRetryCount"`
		} `json:"RestartPolicy"`
		AutoRemove           bool          `json:"AutoRemove"`
		VolumeDriver         string        `json:"VolumeDriver"`
		VolumesFrom          interface{}   `json:"VolumesFrom"`
		CapAdd               interface{}   `json:"CapAdd"`
		CapDrop              interface{}   `json:"CapDrop"`
		DNS                  []interface{} `json:"Dns"`
		DNSOptions           []interface{} `json:"DnsOptions"`
		DNSSearch            []interface{} `json:"DnsSearch"`
		ExtraHosts           interface{}   `json:"ExtraHosts"`
		GroupAdd             interface{}   `json:"GroupAdd"`
		IpcMode              string        `json:"IpcMode"`
		Cgroup               string        `json:"Cgroup"`
		Links                interface{}   `json:"Links"`
		OomScoreAdj          int           `json:"OomScoreAdj"`
		PidMode              string        `json:"PidMode"`
		Privileged           bool          `json:"Privileged"`
		PublishAllPorts      bool          `json:"PublishAllPorts"`
		ReadonlyRootfs       bool          `json:"ReadonlyRootfs"`
		SecurityOpt          interface{}   `json:"SecurityOpt"`
		UTSMode              string        `json:"UTSMode"`
		UsernsMode           string        `json:"UsernsMode"`
		ShmSize              int           `json:"ShmSize"`
		ConsoleSize          []int         `json:"ConsoleSize"`
		Isolation            string        `json:"Isolation"`
		CPUShares            int           `json:"CpuShares"`
		Memory               int           `json:"Memory"`
		NanoCpus             int           `json:"NanoCpus"`
		CgroupParent         string        `json:"CgroupParent"`
		BlkioWeight          int           `json:"BlkioWeight"`
		BlkioWeightDevice    interface{}   `json:"BlkioWeightDevice"`
		BlkioDeviceReadBps   interface{}   `json:"BlkioDeviceReadBps"`
		BlkioDeviceWriteBps  interface{}   `json:"BlkioDeviceWriteBps"`
		BlkioDeviceReadIOps  interface{}   `json:"BlkioDeviceReadIOps"`
		BlkioDeviceWriteIOps interface{}   `json:"BlkioDeviceWriteIOps"`
		CPUPeriod            int           `json:"CpuPeriod"`
		CPUQuota             int           `json:"CpuQuota"`
		CPURealtimePeriod    int           `json:"CpuRealtimePeriod"`
		CPURealtimeRuntime   int           `json:"CpuRealtimeRuntime"`
		CpusetCpus           string        `json:"CpusetCpus"`
		CpusetMems           string        `json:"CpusetMems"`
		Devices              []interface{} `json:"Devices"`
		DiskQuota            int           `json:"DiskQuota"`
		KernelMemory         int           `json:"KernelMemory"`
		MemoryReservation    int           `json:"MemoryReservation"`
		MemorySwap           int           `json:"MemorySwap"`
		MemorySwappiness     int           `json:"MemorySwappiness"`
		OomKillDisable       bool          `json:"OomKillDisable"`
		PidsLimit            int           `json:"PidsLimit"`
		Ulimits              interface{}   `json:"Ulimits"`
		CPUCount             int           `json:"CpuCount"`
		CPUPercent           int           `json:"CpuPercent"`
		IOMaximumIOps        int           `json:"IOMaximumIOps"`
		IOMaximumBandwidth   int           `json:"IOMaximumBandwidth"`
	} `json:"HostConfig"`
	NetworkingConfig struct {
		EndpointsConfig struct {
		} `json:"EndpointsConfig"`
	} `json:"NetworkingConfig"`
}

func handleHTTP(w http.ResponseWriter, req *http.Request, l net.Listener) {

	fd, err := l.(*net.UnixListener).File()
	if err != nil {
		panic(err)
	}

	sockCreds, _ := syscall.GetsockoptUcred(int(fd.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
	fmt.Println(sockCreds.Uid)

	fmt.Printf("Requested : %s\n", req.URL.Path)

	if req.Method == "POST" {

		//Closing the req.Body once we are done reading all of it.
		//Construct a new req.Body to make http length happy
		reqBody, err := ioutil.ReadAll(req.Body)
		defer req.Body.Close()
		req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody))

		if err != nil {
			log.Fatal(err)
		}

		var docker dockerCreateObject
		json.Unmarshal(reqBody, &docker)

		//Get the user POST data
		fmt.Printf("DOCKER USER: %s | OS USER: %d\n", docker.User, sockCreds.Uid)
	}

	u := &httpunix.Transport{
		DialTimeout:           100 * time.Millisecond,
		RequestTimeout:        1 * time.Second,
		ResponseHeaderTimeout: 1 * time.Second,
	}
	u.RegisterLocation("docker-socket", "/var/run/docker.sock")

	req.URL.Scheme = "http+unix"
	req.URL.Host = "docker-socket"

	resp, err := u.RoundTrip(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)

}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func main() {

	socket := "/var/run/docker-socket-go.sock"
	os.Remove(socket)
	unixListener, err := net.Listen("unix", socket)

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { handleHTTP(w, r, unixListener) }),
	}

	if err != nil {
		panic(err)
	}

	log.Fatal(server.Serve(unixListener))

}
