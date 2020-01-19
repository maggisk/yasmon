package main

import "time"

func Configure(register func(string, func(chan string))) string {
	register("network", NetworkComp(5*time.Second))
	register("volume", VolumeComp)
	register("backlight", BacklightComp)
	register("date", DateComp(time.Minute, "Mon Jan _2 15:04"))
	register("keyboard", KeyboardLayoutComp)
	register("power", BatteryPowerComp(2*time.Second, " ", " "))
	register("charge", BatteryChargeComp(time.Minute))

	return " {network}   {backlight}%  墳 {volume}%  {power}{charge}%   {keyboard}  {date}"
}
