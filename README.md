Something to make reading articles on Medium bearable.

Very experimental.

```
go get github.com/PuerkitoBio/goquery
go get github.com/gorilla/mux

go build mitm.go

./mitm

http://127.0.0.1:8888?url=%medium_url%
```