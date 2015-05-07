build:
	docker run --rm -v $(PWD):/usr/src/github.com/BrianBland/palette -w /usr/src/github.com/BrianBland/palette -e 'GOPATH=/usr/src/github.com/BrianBland/palette/Godeps/_workspace:/usr' golang:1.4.2 go build -v './cmd/palette/palette.go'
