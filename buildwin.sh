~/go/bin/go-winres simply --product-version=git-tag --file-version=git-tag --arch=amd64,386 --file-description="donationalerts vtube studio plugin" --product-name="donation-vts" --copyright="2023 Sokolas" --icon="winres/bug.ico"

GOOS=windows GOARCH=amd64 go build -o build64/donations-vts.exe .
GOOS=windows GOARCH=386 go build -o build32/donations-vts.exe .