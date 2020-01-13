package main

import "time"

func Configure(register func(string, func(chan string))) string {
	register("network", NetworkComp(2*time.Second))
	register("volume", VolumeComp)
	register("batt", BatteryComp(5*time.Second, BatteryFormat))
	register("date", DateComp("Mon Jan _2 15:04:05"))

	// KbLayoutComp uses xev (pacman -S xorg-xev) to watch for keyboad layout changes
	// Switch to KbLayoutPollComp(time.Second) to use a polling loop in stead
	register("keyboard", KbLayoutComp)

	return " {network}  墳 {volume}  {batt}   {keyboard}  {date}"
}
