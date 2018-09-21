# go-loadtest

`go get github.com/jelinden/go-loadtest`

`cd $GOPATH/src/github.com/jelinden/go-loadtest`

Change addresses in

```
var listOfAddresses = []string{
	"https://www.uutispuro.fi/fi",
	"https://www.google.fi",
	"https://portfolio.jelinden.fi",
	"https://jelinden.fi",
}
```

`vgo build && ./go-loadtest`
