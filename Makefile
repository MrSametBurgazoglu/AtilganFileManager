.PHONY: build install dist clean

BINARY_NAME=atilgan

build:
	go build -o $(BINARY_NAME) .

install: build
	install -D -m 0755 $(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	install -D -m 0644 io.github.mrsametburgazoglu.AtilganFileManager.desktop /usr/share/applications/io.github.mrsametburgazoglu.AtilganFileManager.desktop
	install -D -m 0644 atilgan_icon.svg /usr/share/icons/hicolor/scalable/apps/io.github.mrsametburgazoglu.AtilganFileManager.svg
	install -D -m 0644 io.github.mrsametburgazoglu.AtilganFileManager.metainfo.xml /usr/share/metainfo/io.github.mrsametburgazoglu.AtilganFileManager.metainfo.xml

dist:
	goreleaser release --snapshot --clean

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/
