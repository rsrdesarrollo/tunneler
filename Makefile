PROJECT=github.com/rsrdesarrollo/tunneler
BINS=tunnelerd tunnelerc

.PHONY: all clean $(BINS)

all: $(BINS)

$(BINS):
	go build -i -o build/$@ -v -ldflags="-X main.version=$$(git describe --tags --long)" $(PROJECT)/main/$@

clean:
	rm $(BINS)

