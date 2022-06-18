package main

import (
	"log"
	"os"
	"sync"

	"file_client/file_downloader"
	"file_client/links_extractor"
)

type channels_to_worker struct {
	comm   chan int
	finish chan int
}

func download_files_earliest_match_occurence(file_list []string, matching_string string, out_dir_path string) {
	var wait_group sync.WaitGroup
	wait_group.Add(len(file_list) + 1)
	file_url_to_channels := make(map[string]channels_to_worker)
	channel_to_mananger := make(chan file_downloader.Channel_to_mananger_type)
	for _, file_url := range file_list {
		channel_to_worker := make(chan int, len(file_list))
		final_channel_to_worker := make(chan int, 1)
		file_url_to_channels[file_url] = channels_to_worker{channel_to_worker, final_channel_to_worker}
		go func(file_url string, channel_to_mananger chan file_downloader.Channel_to_mananger_type, channel_to_worker chan int, final_channel_to_worker chan int, wait_group *sync.WaitGroup) {
			defer wait_group.Done()
			f_dowmloader := file_downloader.NewEarliestMatchFileDownloader(file_url, matching_string, out_dir_path, channel_to_mananger, channel_to_worker, final_channel_to_worker)
			f_dowmloader.Download()
		}(file_url, channel_to_mananger, channel_to_worker, final_channel_to_worker, &wait_group)
	}

	go func(size int, channel_to_mananger chan file_downloader.Channel_to_mananger_type, file_url_to_channels map[string]channels_to_worker, wait_group *sync.WaitGroup) {
		defer wait_group.Done()
		defer close(channel_to_mananger)
		earliest_occurance_position := file_downloader.POSITION_NOT_SET
		ongoing_workers := size
		for ongoing_workers > 0 {
			data_from_worker := <-channel_to_mananger
			if data_from_worker.Position == file_downloader.WORKER_CANCELED_MESSAGE {
				channels := file_url_to_channels[data_from_worker.Id]
				close(channels.comm)
				close(channels.finish)
				delete(file_url_to_channels, data_from_worker.Id)
				log.Printf("Worker %s cancelled", data_from_worker.Id)
				ongoing_workers--
				continue
			}
			if data_from_worker.Position == file_downloader.WORKER_FINISHED_MESSAGE {
				log.Printf("Worker %s finished, waiting for confimation", data_from_worker.Id)
				ongoing_workers--
				continue
			}
			if earliest_occurance_position == file_downloader.POSITION_NOT_SET || data_from_worker.Position < earliest_occurance_position {
				earliest_occurance_position = data_from_worker.Position
				for _, channels_to_worker := range file_url_to_channels {
					channels_to_worker.comm <- earliest_occurance_position
				}
			}
		}
		for _, final_channel_to_worker := range file_url_to_channels {
			final_channel_to_worker.finish <- earliest_occurance_position
		}
		log.Println("Manager finished")
	}(len(file_list), channel_to_mananger, file_url_to_channels, &wait_group)

	wait_group.Wait()
}

func main() {
	if len(os.Args) != 4 {
		log.Fatalf("Expecting server URL, matching string, output dir path as arguments, e.g. %s 'localhost:8080' 'A' ./out", os.Args[0])
	}

	file_server_url := os.Args[1]
	matching_string := os.Args[2]
	output_dir_path := os.Args[3]

	files_urls := links_extractor.From_url(file_server_url)
	download_files_earliest_match_occurence(files_urls, matching_string, output_dir_path)
}
