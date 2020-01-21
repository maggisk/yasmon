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

	// Go wild with our own bash commands (unread email, weather etc.)
	// Pipes are supported, but anything more complex should probably go in a
	// stand-alone bash script
	register("weather", BashComp(time.Hour, `curl http://wttr.in/?format=1`))
	// register("mail", BashComp(10 * time.Minute, `unreadmailcount`))
	// register("status", BashComp(time.Second, `cat ~/somestatusfile`))

	return " {weather}  {network}   {backlight}%  墳 {volume}%  {power}{charge}%   {keyboard}  {date}"
}
