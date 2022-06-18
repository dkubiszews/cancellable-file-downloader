package file_downloader

import (
	"file_client/utils"
	"log"
)

const WORKER_FINISHED_MESSAGE = -1
const WORKER_CANCELED_MESSAGE = -2
const ALL_WORKERS_FINISHED_MESSAGE = -3
const POSITION_NOT_SET = -1

type State int

const (
	STATE_NO_MATCH State = iota
	STATE_LEADING_LOCAL_MATCH
	STATE_LEADING_FOREIGN_MATCH
	STATE_CANCELLED
)

type Channel_to_mananger_type struct {
	Id       string
	Position int
}

type EarliestMatchFileDownloader struct {
	file_url                            string
	out_dir_path                        string
	state                               State
	earliest_occurance_position         int
	earliest_foreign_occurance_position int
	matcher                             utils.DataMatcherWithState
	matching_string_size                int
	channel_to_mananger                 chan Channel_to_mananger_type
	channel_to_worker                   chan int
	final_channel_to_worker             chan int
}

func (downloader *EarliestMatchFileDownloader) evaluate_data(data []byte, position int) bool {
	if downloader.earliest_occurance_position == POSITION_NOT_SET {
		occurence_position, err := downloader.matcher.Match(data)
		if err == nil {
			downloader.earliest_occurance_position = position + occurence_position
		}
	}

	switch downloader.state {
	case STATE_NO_MATCH:
		if downloader.earliest_occurance_position != POSITION_NOT_SET {
			downloader.channel_to_mananger <- Channel_to_mananger_type{downloader.file_url, downloader.earliest_occurance_position}
			downloader.state = STATE_LEADING_LOCAL_MATCH
		}

		if downloader.earliest_foreign_occurance_position != POSITION_NOT_SET {
			if downloader.earliest_occurance_position == POSITION_NOT_SET {
				downloader.state = STATE_LEADING_FOREIGN_MATCH
			} else if downloader.earliest_occurance_position != POSITION_NOT_SET && downloader.earliest_foreign_occurance_position < downloader.earliest_occurance_position {
				downloader.channel_to_mananger <- Channel_to_mananger_type{downloader.file_url, WORKER_CANCELED_MESSAGE}
				downloader.state = STATE_CANCELLED
			}
		}
	case STATE_LEADING_LOCAL_MATCH:
		if downloader.earliest_foreign_occurance_position != POSITION_NOT_SET && downloader.earliest_foreign_occurance_position < downloader.earliest_occurance_position {
			downloader.channel_to_mananger <- Channel_to_mananger_type{downloader.file_url, WORKER_CANCELED_MESSAGE}
			downloader.state = STATE_CANCELLED
		}
	case STATE_LEADING_FOREIGN_MATCH:
		if downloader.earliest_occurance_position != POSITION_NOT_SET && downloader.earliest_occurance_position <= downloader.earliest_foreign_occurance_position {
			downloader.channel_to_mananger <- Channel_to_mananger_type{downloader.file_url, downloader.earliest_occurance_position}
			downloader.state = STATE_LEADING_LOCAL_MATCH
		} else if downloader.earliest_foreign_occurance_position < position {
			downloader.channel_to_mananger <- Channel_to_mananger_type{downloader.file_url, WORKER_CANCELED_MESSAGE}
			downloader.state = STATE_CANCELLED
		}
	default:
		panic("Unknown state. Logic error")
	}
	return downloader.state != STATE_CANCELLED
}

func (downloader *EarliestMatchFileDownloader) Download() {
	f_downloader := NewCancellableFileDownloader(downloader.file_url, downloader.out_dir_path)
	f_downloader.DownloadWithPeek(downloader.matching_string_size, func(data []byte, position int, is_downloading bool) bool {
		result := true
		if is_downloading {
			select {
			case earliest_foreign_occurance_position := <-downloader.channel_to_worker:
				downloader.earliest_foreign_occurance_position = earliest_foreign_occurance_position
				result = downloader.evaluate_data(data, position)
			default:
				result = downloader.evaluate_data(data, position)
			}
		} else {
			downloader.channel_to_mananger <- Channel_to_mananger_type{downloader.file_url, WORKER_FINISHED_MESSAGE}
			if downloader.state == STATE_LEADING_LOCAL_MATCH {
				earliest_foreign_occurance_position := <-downloader.final_channel_to_worker
				if earliest_foreign_occurance_position < downloader.earliest_occurance_position {
					result = false
					log.Printf("Worker %s finished, but found other file with earlier occurance", downloader.file_url)
				} else {
					log.Printf("Worker %s finished and confirmation received. SUCCESS!!!", downloader.file_url)
				}
			} else {
				result = false
				log.Printf("Worker %s finished but wasn't leadin local match", downloader.file_url)
			}
		}
		return result
	})
}

func NewEarliestMatchFileDownloader(file_url string, matching_string string, out_dir_path string, channel_to_mananger chan Channel_to_mananger_type, channel_to_worker chan int, final_channel_to_worker chan int) EarliestMatchFileDownloader {
	return EarliestMatchFileDownloader{file_url, out_dir_path, STATE_NO_MATCH, POSITION_NOT_SET, POSITION_NOT_SET, utils.NewDataMatcherWithState([]byte(matching_string)), len(matching_string), channel_to_mananger, channel_to_worker, final_channel_to_worker}
}
