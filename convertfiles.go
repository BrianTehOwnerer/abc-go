package main

import (
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sync"

	ffmpeg_go "github.com/u2takey/ffmpeg-go"
)

func main() {

	var wg = sync.WaitGroup{}

	//Accepts the first cli argument as the folder
	var foldertoconvertfrom string = os.Args[1]
	var filenameregex = regexp.MustCompile(`(\.\w*$)`)

	var regexforendslash = regexp.MustCompile("/$")
	var regexAAX = regexp.MustCompile(".[a,A]{2}[x,X]$")
	// Checks the folder you pass to the executable and adds a trailing slash if needed
	// If you pass it a . it sets it as current working directory
	if foldertoconvertfrom == "." || len(foldertoconvertfrom) == 0 {
		currentdir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		foldertoconvertfrom = currentdir + "/"
	}
	if !regexforendslash.MatchString(foldertoconvertfrom) {
		foldertoconvertfrom = foldertoconvertfrom + "/"
	}

	//Iterates over a folder and gets all files from it, adding it to file.Name()
	files, err := os.ReadDir(foldertoconvertfrom)
	if err != nil {
		log.Fatal(err)
	}

	//for eatch file in files
	// splits it into two parts, the first being the file name, second is the file extention...//
	for _, file := range files {

		fullfilename := file.Name()
		justFileName := filenameregex.Split(file.Name(), 2)

		wg.Add(1)
		if regexAAX.MatchString(file.Name()) {
			go convertaax(justFileName, foldertoconvertfrom, fullfilename, &wg)
		} else {
			go convertgenericaudio(justFileName, foldertoconvertfrom, fullfilename, &wg)
		}
	}
	wg.Wait()
}

//runs ffmpeg with the given options and coverts from any audio file to an m4a file, with the file extenion m4b
// which is the standard for audio books.

func convertgenericaudio(justFileName []string, foldertoconvertfrom string, fullfilename string, wg *sync.WaitGroup) {
	ffmpeg_go.Input(foldertoconvertfrom+fullfilename).
		Output(foldertoconvertfrom+justFileName[0]+".m4b", ffmpeg_go.KwArgs{"c:a": "aac", "c:v": "copy", "af": "dynaudnorm"}).
		OverWriteOutput().ErrorToStdOut().Run()
	wg.Done()
}

func convertaax(justFileName []string, foldertoconvertfrom string, fullfilename string, wg *sync.WaitGroup) {
	//var activationbytes
	activationbytes, err := ioutil.ReadFile("activation_bytes.txt")
	if err != nil {
		log.Fatal(err)
	}
	ffmpeg_go.Input(foldertoconvertfrom+fullfilename, ffmpeg_go.KwArgs{"activation_bytes": activationbytes}).
		Output(foldertoconvertfrom+justFileName[0]+".m4b", ffmpeg_go.KwArgs{"vn": "", "c:a": "copy"}).
		OverWriteOutput().ErrorToStdOut().Run()
	wg.Done()
}
