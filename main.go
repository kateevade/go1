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
        } else {
            scanner := bufio.NewScanner(resp.Body)
            if resp.StatusCode != http.StatusOK || !scanner.Scan() {
                errorCount++
                resp.Body.Close()
            } else {
                line := strings.TrimSpace(scanner.Text())
                resp.Body.Close()
                data := strings.Split(line, ",")

                if len(data) != 7 {
                    errorCount++
                } else {
                    // Парсим числа
                    vals := make([]float64, 7)
                    bad := false
                    for i, v := range data {
                        num, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
                        if err != nil {
                            bad = true
                            break
                        }
                        vals[i] = num
                    }

                    if bad {
                        errorCount++
                    } else {
                        // Сброс ошибок — данные корректны
                        errorCount = 0

                        loadAvg := vals[0]
                        memTotal := vals[1]
                        memUsage := vals[2]
                        diskTotal := vals[3]
                        diskUsage := vals[4]
                        netTotal := vals[5]
                        netUsage := vals[6]

                        // Load Average
                        if loadAvg > 30 {
                            fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
                        }

                        // Memory usage (>80%)
                        memPercent := (memUsage / memTotal) * 100
                        if memPercent > 80 {
                            fmt.Printf("Memory usage too high: %.0f%%\n", memPercent)
                        }

                        // Disk space (<10% free)
                        diskFree := diskTotal - diskUsage
                        diskFreeMB := diskFree / 1_000_000 // автотест считает МБ как 10^6!

                        if diskFree/diskTotal < 0.1 {
                            fmt.Printf("Free disk space is too low: %.0f Mb left\n", diskFreeMB)
                        }

                        // Network bandwidth usage (>90%)
                        if netUsage/netTotal > 0.9 {
                            netFree := netTotal - netUsage
                            netFreeMbit := netFree / 1_000_000 * 8 // автотест считает Мбит/s через 10^6!
                            fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", netFreeMbit)
                        }
                    }
                }
            }
        }

        if errorCount >= maxErrors {
            fmt.Println("Unable to fetch server statistic")
            return
        }

        time.Sleep(interval)
    }
}
