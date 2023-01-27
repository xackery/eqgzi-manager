NAME ?= eqgzi-manager
VERSION ?= 0.0.1
ICON_PNG ?= icon.png
PACKAGE_NAME ?= com.xackery.eqgzi-manager

run:
	go run main.go
run-mobile:
	go run -tags mobile main.go
build-all: build-darwin build-ios build-linux build-windows build-android
build-darwin:
	@echo "build-darwin: compiling"
	@-mkdir -p bin
	@-rm -rf bin/${NAME}-darwin.zip
	@-rm -rf bin/orcspawn.app
	@time fyne package -os darwin -icon ${ICON_PNG} --tags main.Version=${VERSION}
	@zip -mvr bin/${NAME}-darwin.zip ${NAME}.app -x "*.DS_Store"
build-linux:
	@echo "Building linux"
	@-mkdir -p bin
	@-rm -rf bin/${NAME}-linux
	@time fyne-cross linux -icon ${ICON_PNG}
	@mv fyne-cross/bin/linux-amd64/${NAME} bin/${NAME}-linux
	@-rm -rf fyne-cross/
build-windows:
	@echo "build-windows: compiling"
	@-mkdir -p bin
	@-rm -rf bin/${NAME}-windows
	@time fyne-cross windows -icon ${ICON_PNG}
	@mv fyne-cross/bin/windows-amd64/${NAME}.exe bin/
	@-rm -rf fyne-cross/
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