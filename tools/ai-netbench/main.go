package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type Metrics struct {
	Mode         string    `json:"mode"`
	Rounds       int       `json:"rounds"`
	DataSize     int       `json:"data_size_bytes"`
	Peers        []string  `json:"peers"`
	OpTimes      []float64 `json:"op_times_ms"`
	AvgTime      float64   `json:"avg_time_ms"`
	MinTime      float64   `json:"min_time_ms"`
	MaxTime      float64   `json:"max_time_ms"`
	Jitter       float64   `json:"jitter_ms"`
	BandwidthMBs float64   `json:"bandwidth_MBps"`
	PacketLoss   float64   `json:"packet_loss_pct"`
}

func main() {
	mode := flag.String("mode", "allreduce", "allreduce|ps|bulk")
	peers := flag.String("peers", "", "comma-separated list of peer addresses (host:port)")
	listen := flag.String("listen", ":8080", "address to listen on")
	dataSize := flag.Int("data-size", 1048576, "bytes per message")
	rounds := flag.Int("rounds", 10, "number of collective rounds")
	psRole := flag.String("ps-role", "worker", "worker|server (for ps mode)")
	flag.Parse()

	peerList := []string{}
	if *peers != "" {
		peerList = strings.Split(*peers, ",")
	}

	switch *mode {
	case "allreduce":
		RunAllReduce(*listen, peerList, *dataSize, *rounds)
	case "ps":
		RunParameterServer(*listen, peerList, *dataSize, *rounds, *psRole)
	case "bulk":
		RunBulkTransfer(*listen, peerList, *dataSize)
	default:
		fmt.Fprintf(os.Stderr, "Unknown mode: %s\n", *mode)
		os.Exit(1)
	}
}

// --- AllReduce (Ring) ---
func RunAllReduce(listen string, peers []string, dataSize, rounds int) {
	// Each pod listens and connects to its right neighbor in the ring
	myAddr := getMyAddress(listen)
	allPeers := append(peers, myAddr)
	allPeers = uniqueStrings(allPeers)
	// Sort for deterministic ring
	allPeers = sortStrings(allPeers)
	myIdx := indexOf(myAddr, allPeers)
	if myIdx == -1 {
		log.Fatalf("Could not find my address (%s) in peer list: %v", myAddr, allPeers)
	}
	rightIdx := (myIdx + 1) % len(allPeers)
	rightPeer := allPeers[rightIdx]

	// Start server
	ln, err := net.Listen("tcp", listen)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	defer ln.Close()

	var wg sync.WaitGroup
	var opTimes []float64
	var mu sync.Mutex

	for r := 0; r < rounds; r++ {
		wg.Add(1)
		go func(round int) {
			defer wg.Done()
			// Prepare data
			data := make([]byte, dataSize)
			rand.Read(data)
			start := time.Now()

			// Connect to right neighbor
			conn, err := net.Dial("tcp", rightPeer)
			if err != nil {
				log.Printf("Round %d: dial error: %v", round, err)
				return
			}
			defer conn.Close()
			// Send data
			if _, err := conn.Write(data); err != nil {
				log.Printf("Round %d: write error: %v", round, err)
				return
			}
			// Wait for data from left neighbor
			leftConn, err := ln.Accept()
			if err != nil {
				log.Printf("Round %d: accept error: %v", round, err)
				return
			}
			buf := make([]byte, dataSize)
			_, err = leftConn.Read(buf)
			leftConn.Close()
			if err != nil {
				log.Printf("Round %d: read error: %v", round, err)
				return
			}
			dur := time.Since(start).Seconds() * 1000 // ms
			mu.Lock()
			opTimes = append(opTimes, dur)
			mu.Unlock()
		}(r)
	}
	wg.Wait()
	printMetrics("allreduce", rounds, dataSize, peers, opTimes)
}

// --- Parameter Server ---
func RunParameterServer(listen string, peers []string, dataSize, rounds int, role string) {
	if role == "server" {
		// Server: accept connections, receive data
		ln, err := net.Listen("tcp", listen)
		if err != nil {
			log.Fatalf("PS server listen error: %v", err)
		}
		defer ln.Close()
		var opTimes []float64
		for i := 0; i < rounds*len(peers); i++ {
			conn, err := ln.Accept()
			if err != nil {
				log.Printf("PS server accept error: %v", err)
				continue
			}
			start := time.Now()
			buf := make([]byte, dataSize)
			_, err = conn.Read(buf)
			conn.Close()
			if err != nil {
				log.Printf("PS server read error: %v", err)
				continue
			}
			dur := time.Since(start).Seconds() * 1000
			opTimes = append(opTimes, dur)
		}
		printMetrics("ps-server", rounds, dataSize, peers, opTimes)
	} else {
		// Worker: connect to server, send data
		var opTimes []float64
		for r := 0; r < rounds; r++ {
			for _, server := range peers {
				conn, err := net.Dial("tcp", server)
				if err != nil {
					log.Printf("PS worker dial error: %v", err)
					continue
				}
				data := make([]byte, dataSize)
				rand.Read(data)
				start := time.Now()
				if _, err := conn.Write(data); err != nil {
					log.Printf("PS worker write error: %v", err)
					conn.Close()
					continue
				}
				conn.Close()
				dur := time.Since(start).Seconds() * 1000
				opTimes = append(opTimes, dur)
			}
		}
		printMetrics("ps-worker", rounds, dataSize, peers, opTimes)
	}
}

// --- Bulk Transfer ---
func RunBulkTransfer(listen string, peers []string, dataSize int) {
	// Each pod sends a large blob to all peers
	myAddr := getMyAddress(listen)
	var wg sync.WaitGroup
	var opTimes []float64
	var mu sync.Mutex
	for _, peer := range peers {
		if peer == myAddr {
			continue
		}
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			conn, err := net.Dial("tcp", p)
			if err != nil {
				log.Printf("Bulk dial error: %v", err)
				return
			}
			data := make([]byte, dataSize)
			rand.Read(data)
			start := time.Now()
			if _, err := conn.Write(data); err != nil {
				log.Printf("Bulk write error: %v", err)
				conn.Close()
				return
			}
			conn.Close()
			dur := time.Since(start).Seconds() * 1000
			mu.Lock()
			opTimes = append(opTimes, dur)
			mu.Unlock()
		}(peer)
	}
	wg.Wait()
	printMetrics("bulk", 1, dataSize, peers, opTimes)
}

// --- Helpers ---
func printMetrics(mode string, rounds, dataSize int, peers []string, opTimes []float64) {
	if len(opTimes) == 0 {
		log.Printf("No operations completed, cannot print metrics.")
		return
	}
	min, max, sum := opTimes[0], opTimes[0], 0.0
	for _, t := range opTimes {
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
		sum += t
	}
	avg := sum / float64(len(opTimes))
	jitter := 0.0
	for _, t := range opTimes {
		jitter += (t - avg) * (t - avg)
	}
	jitter = jitter / float64(len(opTimes))
	bw := float64(dataSize) / (avg / 1000.0) / (1024 * 1024) // MB/s
	metrics := Metrics{
		Mode:         mode,
		Rounds:       rounds,
		DataSize:     dataSize,
		Peers:        peers,
		OpTimes:      opTimes,
		AvgTime:      avg,
		MinTime:      min,
		MaxTime:      max,
		Jitter:       jitter,
		BandwidthMBs: bw,
		PacketLoss:   0.0, // Not measured in this simple version
	}
	b, _ := json.MarshalIndent(metrics, "", "  ")
	fmt.Println(string(b))
}

func getMyAddress(listen string) string {
	// Try to get the pod IP from env (K8s Downward API), else use localhost:port
	if ip := os.Getenv("POD_IP"); ip != "" {
		parts := strings.Split(listen, ":")
		port := "8080"
		if len(parts) > 1 {
			port = parts[len(parts)-1]
		}
		return ip + ":" + port
	}
	// Fallback: try to get local IP
	return "localhost" + listen
}

func uniqueStrings(in []string) []string {
	m := map[string]struct{}{}
	out := []string{}
	for _, s := range in {
		if _, ok := m[s]; !ok {
			m[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

func sortStrings(in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	for i := 0; i < len(out)-1; i++ {
		for j := i + 1; j < len(out); j++ {
			if out[i] > out[j] {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
}

func indexOf(s string, arr []string) int {
	for i, v := range arr {
		if v == s {
			return i
		}
	}
	return -1
}
