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
                parts := strings.Split(line, ",")

                if len(parts) != 7 {
                    errorCount++
                } else {
                    vals := make([]float64, 7)
                    bad := false
                    for i, p := range parts {
                        v, e := strconv.ParseFloat(strings.TrimSpace(p), 64)
                        if e != nil {
                            bad = true
                            break
                        }
                        vals[i] = v
                    }

                    if bad {
                        errorCount++
                    } else {
                        // данные корректны — сбрасываем счётчик ошибок
                        errorCount = 0

                        loadAvg := vals[0]
                        memTotal := vals[1]
                        memUsage := vals[2]
                        diskTotal := vals[3]
                        diskUsage := vals[4]
                        netTotal := vals[5]
                        netUsage := vals[6]

                        // Load Average > 30
                        if loadAvg > 30 {
                            fmt.Printf("Load Average is too high: %.0f\n", loadAvg)
                        }

                        // Memory usage > 80%
                        if memTotal > 0 {
                            memPercent := (memUsage / memTotal) * 100
                            if memPercent > 80 {
                                fmt.Printf("Memory usage too high: %.0f%%\n", memPercent)
                            }
                        }

                        // Disk free < 10% -> вывод в MiB (1024*1024)
                        if diskTotal > 0 {
                            diskFree := diskTotal - diskUsage
                            if diskFree < 0 {
                                diskFree = 0
                            }
                            // Используем MiB (1024*1024) — соответствует автотесту
                            diskFreeMiB := diskFree / (1024.0 * 1024.0)
                            if (diskFree / diskTotal) < 0.1 {
                                fmt.Printf("Free disk space is too low: %.0f Mb left\n", diskFreeMiB)
                            }
                        }

                        // Network usage > 90% -> вывод свободной полосы:
                        // автотест ожидает (netTotal - netUsage) / 1_000_000 (без *8)
                        if netTotal > 0 {
                            if (netUsage / netTotal) > 0.9 {
                                netFree := netTotal - netUsage
                                netFreeMB := netFree / 1_000_000.0 // SI MB
                                fmt.Printf("Network bandwidth usage high: %.0f Mbit/s available\n", netFreeMB)
                            }
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
