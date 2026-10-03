package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
	"context"
	"os/signal"
)

type Movie struct{
	ID int `json:"id"`
	Title string `json:"title"`
	Year int `json:"year"`
	Director string  `json:"director"`
}

func main(){
	from := flag.Int("from", 0, "Start movie ID")
	to := flag.Int("to", 0, "End mavie ID")
	workers := flag.Int("workers", 10, "Number of wokers")
	timeout := flag.Int("timeout", 5, "Request timeout")
	flag.Parse()


	switch{
	//Stdout - основной поток, туда полезный резюльтат работы
	//Stderr - поток для ошибок
	//Exit - останавливает программу и выводит ошибку. 
	//Как я поняла можно выводить любую кроме 0, тк 0 - все ок
	case *from == 0:
		fmt.Fprintln(os.Stderr, "Error: Flag '--from' isn't correct.")
		os.Exit(1)
	case *to == 0:
		fmt.Fprintln(os.Stderr, "Error: Flag '--to' isn't correct.")
		os.Exit(1)
	case *from > *to:
		fmt.Fprintln(os.Stderr, "Error: Flag '--from' cannot be greater than '--to'.")
		os.Exit(1)
	case *workers <= 0:
		fmt.Fprintln(os.Stderr, "Error: Flag '--workers' must be graeter than 0.")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel() // Освобождаем ресурсы при выходе из main
	totalJobs := *to - *from + 1
	httpTimeout := time.Duration(*timeout) * time.Second
	jobsChan := make(chan int, totalJobs)
	resultChan := make(chan string, totalJobs)
	for w := 1; w <= *workers; w ++{
		go worker(ctx, jobsChan, resultChan, httpTimeout)
	}
	for id:= *from; id <= *to; id++{jobsChan <- id}
	close(jobsChan)
	for i := 0; i < totalJobs; i++ {
		select{
		case res := <- resultChan:
			fmt.Println(res)
		
		case <-ctx.Done():
			fmt.Println("\nOperation canceled by user (Ctrl+C). Exiting...")
			return
		}
	}
}

func worker(ctx context.Context,
	jobs<- chan int, //<- chan (стрелка слева) можно только читать
	results chan<- string, //chan<- (стрекла справа) можно только записывать
	timeout time.Duration){
		client := &http.Client{Timeout: timeout}
		for id := range jobs{
			// Проверяем контекст перед стартом новой задачи
			if ctx.Err() != nil {return}
			//idText := strconv.Itoa(id)
			siteUrl:= fmt.Sprintf("https://homeworksite.site/%d/info.0.json", id)
			//siteUrl:= "http://127.0.0/" + idText + "/info.0.json"
			req, err := http.NewRequestWithContext(ctx, "GET", siteUrl, nil) 
			if err != nil{
				results <- "Movie ID " + strconv.Itoa(id) + " - Request error"
				continue
			}
			responce, err := client.Do(req)
			if err != nil {
			results <- "Movie ID " + strconv.Itoa(id) + " — Request error (timeout or connection lost)"
			continue
			}
			if responce.StatusCode != http.StatusOK {
			results <- "Movie ID " + strconv.Itoa(id) + " — HTTP Error: " + responce.Status
			responce.Body.Close()
			continue
			}
			body, err := io.ReadAll(responce.Body)//считываем
			//body - []byte
			// response.Status - текст вроде "200 OK" или "404 Not Found".
			//response.StatusCode - число статуса (например, 200 или 404).
			//response.Proto - версия сетевого протокола (например, "HTTP/1.1").
			//response.Header - служебные заголовки (размер файла, тип сервера, дата).
			//responce.Body - текст фильма

			responce.Body.Close()//надо закрывать
			if err != nil{results <- "Movie ID " + strconv.Itoa(id) + " - Read body error"
				continue}
			var film Movie//создаем структуру
			err = json.Unmarshal(body, &film)//собираем все красиво 
			//если нет то ошибка если да то пусто 
			//если & не укажем то film сотрется
			if err != nil{
				results <- "Film's ID" + strconv.Itoa(id) + "- Broked JSON"
				continue
			}
			formated := fmt.Sprintf("%d — %s — %d — %s", film.ID, 
			film.Title, film.Year, film.Director)//s
			results <- formated
		}
}
