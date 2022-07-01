# Audio Book Converter
Audio Book Converter writen in go.

This is the start to yet another audiobook converter.
It takes all the files in a folder, and converts them all to m4b files for itunes and other audiobook managers can use.

useage is simple ./abc-go /folder/to/convert

if the file is an aax encrypted file it will automaticly find your file's checksum and connect to the API below and retrive your activation bytes to decrypt your file.
This project uses the API avaible here https://audible-converter.ml/
