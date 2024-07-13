mkdir bin
rsrc -ico icon.ico -manifest eqgzi-manager.exe.manifest
copy /y eqgzi-manager.exe.manifest bin\eqgzi-manager.exe.manifest
go build -buildmode=pie -ldflags="-s -w" -o eqgzi-manager.exe main.go
move eqgzi-manager.exe bin/eqgzi-manager.exe
cd bin && eqgzi-manager.exe c:\games\eq\rebuildeq\rkp.eqg