package main

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
)

type Download struct {
	link              string
	fileName          string
	destinationFolder string
}

func NewDownload(link string, destinationFolder string) (*Download, error) {

	_, destErr := os.Stat(destinationFolder)

	//If the folder doesn't exist, return null
	if os.IsNotExist(destErr) {
		return nil, destErr
	}

	escapedLink, err := url.QueryUnescape(path.Base(link))
	if err != nil {
		return nil, err
	}
	decodedLink := escapedLink

	dl := &Download{}
	dl.link = link
	dl.fileName = decodedLink
	dl.destinationFolder = destinationFolder

	return dl, nil
}

func (download *Download) DownloadFile() (bool, *string) {
	if !Silent {
		fmt.Printf("%sDownloading file: '%s' to %s'%s'\n", blue, download.fileName, reset, download.destinationFolder)
	}

	req, _ := http.NewRequest("GET", download.link, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		LogErrorIfNecessary("", &err)
		return false, nil
	}

	defer func(Body io.ReadCloser) {
		e1 := Body.Close()
		if e1 != nil {
			LogErrorIfNecessary("", &err)
		}
	}(resp.Body)

	if LogErrorIfNecessary(fmt.Sprintf("Failed to download %s to %s: %s", download.link, download.destinationFolder, err), &err) {
		return false, nil
	}

	partialFileFormat := fmt.Sprintf("%s%s", download.fileName, ".tmp") // it is .tmp, but it can be any extension )
	tempFile := path.Join(os.TempDir(), partialFileFormat)
	f, e2 := os.OpenFile(tempFile, os.O_CREATE|os.O_WRONLY, 0644)

	// We don't care if this errors, as long as the file is gone if something dramatic happens such as a panic
	// Unfortunately the file stagnates in the Temp Folder is the application is interrupted with things like CTRL+C (^C)
	// Or a general computer power failure.
	// New data won't be appended, it is newly written to, so file stagnation leading to file corruption won't be a problem

	if LogErrorIfNecessary(fmt.Sprintf("Failed to create file %s: %s", download.fileName, e2), &e2) {
		return false, nil
	}

	var err2 error
	if Silent {
		_, e3 := io.Copy(f, resp.Body)
		err2 = e3
	} else {
		bar := CreateProgressBar(resp.ContentLength)

		_, e3 := io.Copy(io.MultiWriter(f, bar), resp.Body)
		err2 = e3
	}

	if LogErrorIfNecessary(fmt.Sprintf("Failed to copy file %s: %s", download.fileName, err2), &err2) {
		return false, nil
	}

	defer f.Close()
	defer os.Remove(tempFile)

	destinationPath := path.Join(download.destinationFolder, download.fileName)
	//outputFile, e5 := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY, 0666)
	//if e5 != nil {
	//	LogErrorIfNecessary("Unable to create a destination file", &e5)
	//	return false
	//}

	bytes, e6 := ioutil.ReadFile(tempFile)
	if e6 != nil {
		LogErrorIfNecessary("Unable to read the temporary file", &e6)
		return false, nil
	}
	copyFail := ioutil.WriteFile(destinationPath, bytes, 0644)
	f.Close()
	os.Remove(tempFile)
	//_, copyFail := io.Copy(outputFile, f)
	if copyFail != nil {
		LogErrorIfNecessary("Unable to copy the file", &copyFail)
		return false, nil
	}

	if !Silent {
		fmt.Printf("%sComplete: %s%s\n", blue, download.fileName, reset)
	}
	return true, &destinationPath
}

func CreateProgressBar(maxBytes int64) *progressbar.ProgressBar {
	desc := "Downloading"

	bar := progressbar.DefaultBytes(maxBytes, desc)
	return bar
}

// TryDownload Returns Path if any
func TryDownload(link string) *string {
	if !download {
		return nil
	}

	downloadPtr, err := NewDownload(link, cachedFolderLocation)
	if err != nil {
		LogErrorIfNecessary(fmt.Sprintf("Skipping Download of %s", link), &err)
		return nil
	}

	currentLink := downloadPtr.link

	escapedLink, err := url.QueryUnescape(path.Base(currentLink))
	if err != nil {
		LogErrorIfNecessary(fmt.Sprintf("Skipping Download of %s", currentLink), &err)
		return nil
	}

	//LogProgressIfNecessary(fmt.Sprintf("Starting download of %s", escapedLink))
	downloaded, downloadPath := downloadPtr.DownloadFile()
	if !downloaded { // Finish is already logged inside the download function
		LogProgressIfNecessary(fmt.Sprintf("Failed to download %s", escapedLink))
		return nil
	}
	return downloadPath
}

func LogProgressIfNecessary(progress string) {
	if !Silent {
		fmt.Println(progress)
	}
}
