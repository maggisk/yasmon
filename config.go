package main

import "time"

func Configure(register RegFunc) string {
	register("network", NetworkComp)
	register("volume", VolumeComp)
	register("batt", BatteryComp(1 * time.Second, BatteryFormat))
	register("keyboard", KbLayoutComp(time.Second))
	register("date", DateComp("Mon Jan _2 15:04:05"))

	return " {network}  墳 {volume}  {batt}   {keyboard}  {date}"
}
