package main

import "time"

func Configure(register func(string, func(chan string))) string {
	register("volume", VolumeComp(SimpleVolumeFormatter("ﱝ", "奄", "奔", "墳")))
	register("backlight", BacklightComp)
	register("date", DateComp(time.Minute, "Mon Jan _2 15:04"))
	register("keyboard", KeyboardLayoutComp)
	register("power", BatteryPowerComp(2*time.Second, " ", " "))
	register("charge", BatteryChargeComp(time.Minute))
	register("network", NetworkComp(5*time.Second))
	// or if you want to save space and just know if you are connected or not
	// register("network", HasNetwork(time.Second, "", ""))

	// Go wild with our own bash commands (unread email, weather etc.)
	// Pipes are supported, but anything more complex should probably go in a
	// stand-alone bash script
	register("weather", NeedsNetwork(BashComp(10*time.Minute, `curl wttr.in/?format=1`)))
	// register("status", BashComp(time.Second, `cat ~/somestatusfile`))

	return " {weather}  {network}   {backlight}%  {volume}%  {power}{charge}%   {keyboard}  {date}"
}
