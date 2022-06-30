# abc-go
Audio Book Converter writen in go
This is the start to yet another audiobook converter.
It takes all the files in a folder, and converts them all to m4b files for itunes and other audiobook managers can use.

useage is simple ./abc-go /folder/to/convert

if the file is an aax file it will require your audible activation bytes. The program automaticly reads the file activation_bytes.txt from the executable folder.
To get your activation bytes I recomend https://audible-converter.ml/
In the future it will automaticly get these bytes and save them for you.
