all:
	go build tmp.go
release:
	go mod tidy # update all the dependencies
	go build -o centi -ldflags "-s -w" main.go && chmod +x centi
clean:
	rm main
docker:
	mkdir ~/.centi
	docker run -v ~/.centi:/.centi centi
