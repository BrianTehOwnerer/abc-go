package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
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
	var folder string
	var getchecksum, getactivationbytes, recursive, deletefiles bool
	flag.StringVar(&folder, "folder", ".", "Specify the name of the folder you wish to convert.")
	flag.BoolVar(&getchecksum, "checksum", false, "Get the checksum of all AAX audible Files in a directory. Does not convert audio files.")
	flag.BoolVar(&getactivationbytes, "activationbytes", false, "Get the activation bytes of your audible AAX files. Does not convert any audio files.")
	flag.BoolVar(&recursive, "recursive", false, "Recursivly go through sub folders, default is false.")
	flag.BoolVar(&deletefiles, "deletefiles", false, "Delete files once the conversion is complete. Understand, this does not care if the file was a lost pet, or an mp3 file. if its in the folder it WILL be deleted.")
	flag.Parse()
	//Accepts the first cli argument as the folder
	var foldertoconvertfrom string = folder
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
	// but only if we did not get asked for activationbytes or checksum value, if they were true skip this section
	if !getchecksum && !getactivationbytes {
		for _, file := range files {

			fullfileName := file.Name()
			justFileName := filenameregex.Split(file.Name(), 2)

			wg.Add(1)
			if regexAAX.MatchString(file.Name()) {
				activationkey := getactivationkey(fullfileName, foldertoconvertfrom)
				go convertaax(justFileName, foldertoconvertfrom, fullfileName, deletefiles, &wg, activationkey)
			} else {
				go convertgenericaudio(justFileName, foldertoconvertfrom, fullfileName, deletefiles, &wg)
			}
		}
		wg.Wait()
	}
	if getactivationbytes || getchecksum {
		for _, file := range files {

			fullFileName := file.Name()
			if regexAAX.MatchString(file.Name()) {
				checksum := getaaxchecksum(fullFileName, foldertoconvertfrom)
				if getchecksum {
					fmt.Println(checksum)

				}
				if getactivationbytes {
					activaitonkey := getactivationkey(fullFileName, foldertoconvertfrom)
					fmt.Println(activaitonkey)

				}
			}
		}
	}
}

//runs ffmpeg with the given options and coverts from any audio file to an m4a file,
// with the file extenion m4b
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

func convertgenericaudio(justFileName []string, foldertoconvertfrom string, fullfilename string, deletefiles bool, wg *sync.WaitGroup) {
	ffmpeg_go.Input(foldertoconvertfrom+fullfilename).
		Output(foldertoconvertfrom+justFileName[0]+".m4b", ffmpeg_go.KwArgs{"c:a": "aac", "c:v": "copy", "af": "dynaudnorm"}).
		OverWriteOutput().Run()
	if deletefiles == true {
		os.Remove(foldertoconvertfrom + fullfilename)
	}
	wg.Done()
}

func convertaax(justFileName []string, foldertoconvertfrom string, fullfilename string, deletefiles bool, wg *sync.WaitGroup, activationbytes string) {
	ffmpeg_go.Input(foldertoconvertfrom+fullfilename, ffmpeg_go.KwArgs{"activation_bytes": activationbytes}).
		Output(foldertoconvertfrom+justFileName[0]+".m4b", ffmpeg_go.KwArgs{"vn": "", "c:a": "copy"}).
		OverWriteOutput().Run()
	if deletefiles == true {
		os.Remove(foldertoconvertfrom + fullfilename)
	}
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
	resp.Header.Set("User-Agent:", "abc-go/1.1")
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
