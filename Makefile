PROJECT=github.com/rsrdesarrollo/tunneler
BINS=tunnelerd tunnelerc

.PHONY: all clean deps $(BINS)

all: $(BINS)

$(BINS): 
	go build -i -v -o $@.out -ldflags="-X main.version=$$(git describe --tags --long)" $(PROJECT)/main/$@

clean:
	rm $(BINS)

