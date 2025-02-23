
## Build

```go build```

```go build -o owldb```

Assuming you have a file "document.json" that holds your desired
document schema and a file "tokens.json" that holds a set of tokens,
then you could run your program like so:

```./owldb -s document.json -t tokens.json -p 3318```

Note that you can always run your program without building it first as
follows:

```go run main.go -s document.json -t tokens.json -p 3318```

