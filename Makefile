NAME ?= eqgzi-manager
VERSION ?= 0.0.3
ICON_PNG ?= icon.png
PACKAGE_NAME ?= com.xackery.eqgzi-manager

run:
	go run main.go
run-mobile:
	go run -tags mobile main.go

bundle:
	fyne bundle --package client -name blenderIcon assets/blender.svg > client/bundle.go
	fyne bundle --package client -name baseBlend --append assets/base.blend >> client/bundle.go
	fyne bundle --package client -name convertText --append assets/convert.bat >> client/bundle.go
	fyne bundle --package client -name copyEQText --append assets/copy_eq.bat >> client/bundle.go
	fyne bundle --package client -name copyServerText --append assets/copy_server.bat >> client/bundle.go
	fyne bundle --package client -name eqIcon --append assets/eq.svg >> client/bundle.go
	fyne bundle --package client -name whitePng --append assets/white.png >> client/bundle.go
	echo ${VERSION} > "assets/version.txt"
	fyne bundle --package client -name VersionText --append assets/version.txt >> client/bundle.go
build-all: build-darwin build-ios build-linux build-windows build-android
build-darwin:
	@echo "build-darwin: compiling"
	@-mkdir -p bin
	@-rm -rf bin/${NAME}-darwin.zip
	@-rm -rf bin/orcspawn.app
	@time fyne package -os darwin -icon ${ICON_PNG} --appVersion ${VERSION} --tags main.Version=${VERSION}
	@zip -mvr bin/${NAME}-${VERSION}-darwin.zip ${NAME}.app -x "*.DS_Store"
build-linux:
	@echo "Building linux"
	@-mkdir -p bin
	@-rm -rf bin/${NAME}-linux
	@time fyne-cross linux -icon ${ICON_PNG}
	@mv fyne-cross/bin/linux-amd64/${NAME} bin/${NAME}-linux
	@-rm -rf fyne-cross/
build-windows:
	@echo "build-windows: compiling"
	-mkdir -p bin
	-rm bin/${NAME}-*-windows.zip
	fyne-cross windows -icon ${ICON_PNG}
	mv fyne-cross/bin/windows-amd64/${NAME}.exe bin/
	-rm -rf fyne-cross/
	cd bin && zip -mv ${NAME}-${VERSION}-windows.zip ${NAME}.exe
build-ios:
	@echo "build-ios: compiling"
	@-mkdir -p bin
	@-rm -rf bin/${NAME}-ios.zip
	@DISABLE_MANUAL_TARGET_ORDER_BUILD_WARNING=1 time fyne package -os ios -appID ${PACKAGE_NAME} -icon ${ICON_PNG}
	@zip -mvr bin/${NAME}-ios.zip ${NAME}.app -x "*.DS_Store"
build-android:
	@echo "build-android: compiling"
	@-mkdir -p bin
	@-rm bin/${NAME}.apk
	@ANDROID_NDK_HOME=~/Library/Android/sdk/ndk-bundle fyne package -os android -appID ${PACKAGE_NAME} -icon ${ICON_PNG}
	@mv ${NAME}.apk bin/${NAME}.apk
build-web:
	@echo "build-web: compiling"
	@-mkdir -p bin
	@#-rm -rf bin/${NAME}-darwin.zip
	@time fyne package -os web -icon ${ICON_PNG}
	@#zip -mvr bin/${NAME}-darwin.zip ${NAME}.app -x "*.DS_Store"