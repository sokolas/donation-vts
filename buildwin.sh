mkdir build32
mkdir build64
GOOS=windows GOARCH=amd64 go build -o build64/donations-vts.exe .
GOOS=windows GOARCH=386 go build -o build32/donations-vts.exe .