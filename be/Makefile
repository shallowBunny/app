build:
	go build -o shallowBunny cmd/main.go 

run:		
	go run cmd/main.go

duck:
	reflex -r '\.(go|yaml)$$' -s -- go run cmd/main.go -config=configs/duck.yaml

duckdev:
	reflex -r '\.(go|yaml)$$' -s -- go run cmd/main.go -config=configs/duckdev.yaml

bunny:
	reflex -r '\.(go|yaml)$$' -s -- go run cmd/main.go -config=configs/bunny.yaml

demo:
	reflex -r '\.(go|yaml)$$' -s -- go run cmd/main.go -config=configs/demo.yaml
