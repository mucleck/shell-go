package main

func main() {
	if err := ExecuteShell(); err != nil {
		panic(err)
	}
}
