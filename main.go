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
	interval  = 1 * time.Second
	maxErrors = 3
)

func main() {
	errorCount := 0

	for {
		resp, err := http.Get(serverURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			errorCount++
			if resp != nil {
				resp.Body.Close()
			}
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(interval)
			continue
		}

		scanner := bufio.NewScanner(resp.Body)
		if !scanner.Scan() {
			errorCount++
			resp.Body.Close()
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(interval)
			continue
		}

		line := strings.TrimSpace(scanner.Text())
		resp.Body.Close()

		fields := strings.Split(line, ",")
		if len(fields) != 7 {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(interval)
			continue
		}

		vals := make([]float64, 7)
		valid := true
		for i, f := range fields {
			v, err := strconv.ParseFloat(strings.TrimSpace(f), 64)
			if err != nil {
				valid = false
				break
			}
			vals[i] = v
		}
		if !valid {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic")
				return
			}
			time.Sleep(interval)
			continue
		}

		// Сброс ошибок после успешного запроса
		errorCount = 0

		loadAvg := vals[0]
		memTotal := vals[1]
		memUsage := vals[2]
		diskTotal := vals[3]
		diskUsage := vals[4]
		netTotal := vals[5]
		netUsage := vals[6]

		// LOAD
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
		}

		// MEMORY
		if memTotal > 0 {
			memPercent := int(memUsage * 100 / memTotal)
			if memPercent > 80 {
				fmt.Printf("Memory usage too high: %d%%\n", memPercent)
			}
		}

		// DISK
		diskFree := diskTotal - diskUsage
		if diskFree < diskTotal*0.1 {
			diskFreeMB := int64(diskFree) / 1024 / 1024
			fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMB)
		}

		// NETWORK
		if netUsage > netTotal*0.9 {
			netFree := netTotal - netUsage
			netFreeMbit := int64(netFree) / 1_000_000
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netFreeMbit)
		}

		time.Sleep(interval)
	}
}
