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
                    vals := make([]int64, 7)
                    bad := false

                    for i, v := range data {
                        f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
                        if err != nil {
                            bad = true
                            break
                        }
                        vals[i] = int64(f)
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

                        // LOAD AVERAGE
                        if loadAvg > 30 {
                            fmt.Printf("Load Average is too high: %d\n", loadAvg)
                        }

                        // MEMORY
                        if memUsage*100/memTotal > 80 {
                            fmt.Printf("Memory usage too high: %d%%\n", memUsage*100/memTotal)
                        }

                        // DISK (truncate)
                        diskFree := diskTotal - diskUsage
                        diskFreeMB := diskFree / 1_000_000

                        if diskFree*10 < diskTotal {
                            fmt.Printf("Free disk space is too low: %d Mb left\n", diskFreeMB)
                        }

                        // NETWORK (truncate)
                        if netUsage*10 > netTotal*9 {
                            netFree := netTotal - netUsage
                            netFreeMbit := (netFree * 8) / 1_000_000
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
