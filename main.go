package main

func main() {
	err := startStreaming("localhost", 56565)
	if err != nil {
		panic(err)
	}
}
