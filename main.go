
go
                
package main

import (

    "fmt"

    "io/ioutil"

    "net/http"

    "strconv"

    "strings"

    "time"

)

const (

    serverURL = "http://srv.msk01.gigacorp.local/_stats"

    interval  = 5 * time.Second // Интервал опроса сервера

    maxErrors = 3 // Максимальное количество допустимых ошибок

)

func main() {

    errorCount := 0

    for {

        resp, err := http.Get(serverURL)

        if err != nil {

            errorCount++

            fmt.Println("Ошибка при запросе:", err)

            if errorCount >= maxErrors {

                fmt.Println("Невозможно получить статистику сервера.")

                return

            }

            time.Sleep(interval)

            continue

        }

        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {

            errorCount++

            fmt.Printf("Неверный статус код: %d\n", resp.StatusCode)

            if errorCount >= maxErrors {

                fmt.Println("Невозможно получить статистику сервера.")

                return

            }

            time.Sleep(interval)

            continue

        }

        body, err := ioutil.ReadAll(resp.Body)

        if err != nil {

            errorCount++

            fmt.Println("Ошибка чтения тела ответа:", err)

            if errorCount >= maxErrors {

                fmt.Println("Невозможно получить статистику сервера.")

                return

            }

            time.Sleep(interval)

            continue

        }

        data := strings.Split(string(body), ",")

        if len(data) != 7 {

            errorCount++

            fmt.Println("Неверный формат данных")

            if errorCount >= maxErrors {

                fmt.Println("Невозможно получить статистику сервера.")

                return

            }

            time.Sleep(interval)

            continue

        }

        errorCount = 0 // Сброс счетчика ошибок при успешном получении данных

        loadAvg, _ := strconv.ParseFloat(data[0], 64)

        memTotal, _ := strconv.ParseFloat(data[1], 64)

        memUsage, _ := strconv.ParseFloat(data[2], 64)

        diskTotal, _ := strconv.ParseFloat(data[3], 64)

        diskUsage, _ := strconv.ParseFloat(data[4], 64)

        netTotal, _ := strconv.ParseFloat(data[5], 64)

        netUsage, _ := strconv.ParseFloat(data[6], 64)

        if loadAvg > 30 {

            fmt.Printf("Load Average is too high: %.2f\n", loadAvg)

        }

        memPercent := (memUsage / memTotal) * 100

        if memPercent > 80 {

            fmt.Printf("Memory usage too high: %.2f%%\n", memPercent)

        }

        diskFreeMb := (diskTotal - diskUsage) / (1024 * 1024)

        if diskFreeMb < 0 {

            diskFreeMb = 0

        }

        if (diskTotal - diskUsage) / diskTotal < 0.1 {

            fmt.Printf("Free disk space is too low: %.2f Mb left\n", diskFreeMb)

        }

        netAvailableMbit := ((netTotal - netUsage) / 1024 / 1024) * 8

        if (netUsage / netTotal) > 0.9 {

            fmt.Printf("Network bandwidth usage high: %.2f Mbit/s available\n", netAvailableMbit)

        }

        time.Sleep(interval)

    }

}

            