#!/bin/bash

if [ -z "$1" ]
  then
    echo "≫ Error: enter a version, e.g. 1.0"
    exit 1
fi
version=$1

rm -rf build

# Windows
echo "≫ Building Windows binaries..."
GOOS=windows GOARCH=arm64 go build -o build/tn-cli.exe .
cd build
tar -czvf "tn-cli-$version-windows-arm64.tar.gz" tn-cli.exe &> /dev/null
rm tn-cli.exe
cd ..
GOOS=windows GOARCH=amd64 go build -o build/tn-cli.exe .
cd build
tar -czvf "tn-cli-$version-windows-x86_64.tar.gz" tn-cli.exe &> /dev/null
rm tn-cli.exe
cd ..
GOOS=windows GOARCH=386 go build -o build/tn-cli.exe .
cd build
tar -czvf "tn-cli-$version-windows-386.tar.gz" tn-cli.exe &> /dev/null
rm tn-cli.exe
cd ..

# macOS - Darwin
echo "≫ Building macOS - Darwin binaries..."
GOOS=darwin GOARCH=arm64 go build -o build/tn-cli .
cd build
tar -czvf "tn-cli-$version-darwin-arm64.tar.gz" tn-cli &> /dev/null
rm tn-cli
cd ..
GOOS=darwin GOARCH=amd64 go build -o build/tn-cli .
cd build
tar -czvf "tn-cli-$version-darwin-x86_64.tar.gz" tn-cli &> /dev/null
rm tn-cli
cd ..

# Linux
echo "≫ Building Linux binaries..."
GOOS=linux GOARCH=arm64 go build -o build/tn-cli .
cd build
tar -czvf "tn-cli-$version-linux-arm64.tar.gz" tn-cli &> /dev/null
rm tn-cli
cd ..
GOOS=linux GOARCH=amd64 go build -o build/tn-cli .
cd build
tar -czvf "tn-cli-$version-linux-x86_64.tar.gz" tn-cli &> /dev/null
rm tn-cli
cd ..
GOOS=linux GOARCH=386 go build -o build/tn-cli .
cd build
tar -czvf "tn-cli-$version-linux-386.tar.gz" tn-cli &> /dev/null
rm tn-cli
cd ..

echo "Done."