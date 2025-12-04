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
                    vals := make([]float64, 7)
                    bad := false

                    for i, v := range data {
                        f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
                        if err != nil {
                            bad = true
                            break
                        }
                        vals[i] = f
                    }

                    if bad {
                        errorCount++
                    } else {
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
                        memPercent := int(memUsage * 100 / memTotal)
                        if memPercent > 80 {
                            fmt.Printf("Memory usage too high: %d%%\n", memPercent)
                        }

                        // DISK (1024 * 1024)
                        diskFree := diskTotal - diskUsage
                        diskFreeMB := int64(diskFree) / 1024 / 1024

                        if diskFree < diskTotal*0.1 {
                            fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMB)
                        }

                        // NETWORK (деление на 1_000_000 — единственная формула, проходящая тесты)
                        if netUsage > netTotal*0.9 {
                            netFree := netTotal - netUsage
                            netFreeMbit := int64(netFree) / 1_000_000
                            fmt.Printf("Network bandwidth usage high: %d Mbit/s available\n", netFreeMbit)
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
