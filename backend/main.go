package main

import "mncet/mncet/mncet"

func main() {
	if mncet.InitStart() {
		mncet.Start()
	}

}
