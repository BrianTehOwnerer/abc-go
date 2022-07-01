package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

	// This gets your personal activation bytes from the file activation_bytes.txt file from the same
	// folder as your program folder. this is legacy code, now it gets the activation code automaticly
	/*	activaitonfile, _ := os.Open("activation_bytes.txt")
		reader := bufio.NewReader(activaitonfile)
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			log.Fatal(err)
		}
		activationbytes := line
	*/

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

		fullfileName := file.Name()
		justFileName := filenameregex.Split(file.Name(), 2)

		wg.Add(1)
		if regexAAX.MatchString(file.Name()) {
			activationkey := getactivationkey(fullfileName, foldertoconvertfrom)
			go convertaax(justFileName, foldertoconvertfrom, fullfileName, &wg, activationkey)
		} else {
			go convertgenericaudio(justFileName, foldertoconvertfrom, fullfileName, &wg)
		}
	}
	wg.Wait()
}

//runs ffmpeg with the given options and coverts from any audio file to an m4a file, with the file extenion m4b
// which is the standard for audio books.

func getaaxchecksum(fullFilename string, foldertoconvertfrom string) string {
	aaxfile, err := os.Open(foldertoconvertfrom + fullFilename)
	if err != nil {
		fmt.Println("error opening file")
	}
	//this sets the size of checksumbytes to 20 bytes long (which is the size of the aax file checksum)
	checksumbytes := make([]byte, 20)

	//jumps to the 653rd byte of the aax file and reads in the next 20 bytes
	aaxfile.Seek(653, 0)
	aaxfile.Read(checksumbytes)

	//takes the raw binarydata and converts it to hex encoding
	var checksum string = string(hex.EncodeToString(checksumbytes))
	return checksum
}

func convertgenericaudio(justFileName []string, foldertoconvertfrom string, fullfilename string, wg *sync.WaitGroup) {
	ffmpeg_go.Input(foldertoconvertfrom+fullfilename).
		Output(foldertoconvertfrom+justFileName[0]+".m4b", ffmpeg_go.KwArgs{"c:a": "aac", "c:v": "copy", "af": "dynaudnorm"}).
		OverWriteOutput().ErrorToStdOut().Run()
	wg.Done()
}

func convertaax(justFileName []string, foldertoconvertfrom string, fullfilename string, wg *sync.WaitGroup, activationbytes string) {
	ffmpeg_go.Input(foldertoconvertfrom+fullfilename, ffmpeg_go.KwArgs{"activation_bytes": activationbytes}).
		Output(foldertoconvertfrom+justFileName[0]+".m4b", ffmpeg_go.KwArgs{"vn": "", "c:a": "copy"}).
		OverWriteOutput().ErrorToStdOut().Run()
	wg.Done()
}

//this function takes the checksum we found in getaaxchecksum and makes an API call to get the decryption bytes
func getactivationkey(fullFileName string, foldertoconvertfrom string) string {
	checksum := getaaxchecksum(fullFileName, foldertoconvertfrom)
	resp, err := http.Get("https://aax.api.j-kit.me/api/v2/activation/" + checksum)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()
	resp.Header.Set("User-Agent:", "abc-go/1.0")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//reads in the https responce data and makes a maped string list
	var apiresponce map[string]interface{}
	json.Unmarshal(body, &apiresponce)
	//sets activationkey to the responded api's "activationBytes"
	activationkey := fmt.Sprintf("%v", apiresponce["activationBytes"])

	return activationkey

}
