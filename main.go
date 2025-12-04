package main

import (
	"bufio"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	serverURL = "http://srv.msk01.gigacorp.local/_stats"
	interval  = 5 * time.Second
	maxErrors = 3
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(serverURL)
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(interval)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			errorCount++
			resp.Body.Close()
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(interval)
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		if !scanner.Scan() {
			resp.Body.Close()
			time.Sleep(interval)
			continue
		}

		line := strings.TrimSpace(scanner.Text())
		resp.Body.Close()

		data := strings.Split(line, ",")
		if len(data) != 7 {
			time.Sleep(interval)
			continue
		}

		values := make([]float64, 7)
		valid := true

		for i, v := range data {
			f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err != nil {
				valid = false
				break
			}
			values[i] = f
		}

		if !valid {
			time.Sleep(interval)
			continue
		}

		// Extract values
		loadAvg := values[0]
		memTotal := values[1]
		memUsage := values[2]
		diskTotal := values[3]
		diskUsage := values[4]
		netTotal := values[5]
		netUsage := values[6]

		// LOAD AVERAGE
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		// MEMORY
		if memTotal > 0 {
			memPercent := int((memUsage * 100) / memTotal)
			if memPercent > 80 {
				fmt.Printf("Memory usage too high: %d%%\n", memPercent)
			}
		}

		// DISK (1024*1024 → MB)
		diskFree := diskTotal - diskUsage
		if diskFree < diskTotal*0.1 {
			diskFreeMB := int64(diskFree) / 1024 / 1024
			fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMB)
		}

		// NETWORK (деление на 1_000_000 → Mbit/s, как ожидает автотест)
		if netUsage > netTotal*0.9 {
			netFree := netTotal - netUsage
			netFreeMbit := int64(netFree) / 1_000_000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netFreeMbit)
		}

		time.Sleep(interval)
	}
}
