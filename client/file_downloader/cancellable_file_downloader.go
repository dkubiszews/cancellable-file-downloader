package file_downloader

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
)

type CancellableFileDownloader struct {
	file_url     string
	out_dir_path string
}

func extract_filename_from_url_string(url_string string) string {
	file_url, err := url.Parse(url_string)
	if err != nil {
		panic(err)
	}

	return strings.TrimLeft(file_url.Path, "/")
}

func NewCancellableFileDownloader(file_url string, out_dir_path string) CancellableFileDownloader {
	return CancellableFileDownloader{file_url, out_dir_path}
}

func (downloader *CancellableFileDownloader) DownloadWithPeek(chunk_size int, peekFunc func(data []byte, position int, is_downloading bool) bool) {
	filename := extract_filename_from_url_string(downloader.file_url)
	out_filepath := path.Join(downloader.out_dir_path, filename)
	tmp_filepath := out_filepath + ".tmp"
	tmp_file, err := os.Create(tmp_filepath)
	if err != nil {
		log.Fatal(err)
	}

	http_client := http.Client{}
	response, err := http_client.Get(downloader.file_url)
	if err != nil {
		log.Fatal(err)
	}
	defer response.Body.Close()

	buffer := make([]byte, chunk_size)
	completed := false
	cancelled := false
	position := 0
	for !completed && !cancelled {
		size, err := response.Body.Read(buffer)
		if err != nil {
			completed = true
		}

		cancelled = !peekFunc(buffer[:size], position, true)
		tmp_file.Write(buffer[:size])
		position += size
	}

	tmp_file.Close()
	if !cancelled && peekFunc(buffer, position, false) {
		os.Rename(tmp_filepath, out_filepath)
	} else {
		os.Remove(tmp_filepath)
	}
}
