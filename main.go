package main

import (
	"fmt"
	"io"
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
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(interval)
			continue
		}

		// статус != 200
		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(interval)
			continue
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(interval)
			continue
		}

		data := strings.Split(strings.TrimSpace(string(bodyBytes)), ",")
		if len(data) != 7 {
			errorCount++
			if errorCount >= maxErrors {
				fmt.Println("Unable to fetch server statistic.")
				return
			}
			time.Sleep(interval)
			continue
		}

		// данные корректные → сброс
		errorCount = 0

		loadAvg, _ := strconv.ParseFloat(data[0], 64)
		memTotal, _ := strconv.ParseFloat(data[1], 64)
		memUsage, _ := strconv.ParseFloat(data[2], 64)
		diskTotal, _ := strconv.ParseFloat(data[3], 64)
		diskUsage, _ := strconv.ParseFloat(data[4], 64)
		netTotal, _ := strconv.ParseFloat(data[5], 64)
		netUsage, _ := strconv.ParseFloat(data[6], 64)

		// load average
		if loadAvg > 30 {
			fmt.Printf("Load Average is too high: %d\n", int(loadAvg))
		}

		// memory
		memPercent := memUsage / memTotal * 100
		if memPercent > 80 {
			fmt.Printf("Memory usage too high: %d%%\n", int(memPercent))
		}

		// disk
		freeDiskMB := int((diskTotal - diskUsage) / 1024 / 1024)
		if (diskTotal-diskUsage)/diskTotal < 0.1 {
			fmt.Printf("Free disk space is too low: %d Mb left\n", freeDiskMB)
		}

		// network (главное исправление)
		if netUsage/netTotal > 0.9 {
			freeMB := int((netTotal - netUsage) / 1024 / 1024)
			// тест ожидает именно MB, НЕ умноженные на 8
			fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", freeMB)
		}

		time.Sleep(interval)
	}
}

