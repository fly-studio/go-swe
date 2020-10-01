package main

import (
	"fmt"
	"go-swe/astro"
	"time"
)

func main() {

	long, _ := astro.StringToDegrees("116°23'")
	lat, _ := astro.StringToDegrees("39°54'")
	fmt.Printf("Geo: %f %f\n", long, lat)

	geo := &astro.GeographicCoordinates{
		Longitude: astro.ToRadians(long),
		Latitude:  astro.ToRadians(lat),
	}
	tz, _ := time.LoadLocation("Asia/Shanghai")

	year, month, day := time.Now().Date()

	astronomy := astro.NewAstronomy()

	// 节气
	solarTerms, err := astronomy.SolarTerms(year)
	if err != nil {
		fmt.Printf("SolarTerms Error: %s", err.Error())
	}
	for i, jd := range solarTerms {
		fmt.Printf("%s: %v\n", astro.SolarTermsString[i], jd.ToTime(nil).In(tz))
	}

	t := time.Date(year, month, day, 4, 0, 0, 0, time.UTC)
	jd := astro.TimeToJulianDay(t)
	deltaT := astro.DeltaT(jd)
	et := jd.ToEphemerisTime(deltaT)
	etT := et.ToTime(time.UTC)
	fmt.Printf("JD: %f at %v \n", jd, jd.ToTime(time.UTC))
	fmt.Printf("ET: %f at %v deltaT: %v\n", et, etT, deltaT)

	// 太阳
	sunTimes, err := astronomy.SunTwilight(jd, geo, false)
	if err != nil {
		fmt.Printf("SunTwilight Error: %s", err.Error())
	}
	fmt.Printf("Sun Rise: %v\n", sunTimes.Rise.ToTime(nil).In(tz))
	fmt.Printf("Sun Set: %v\n", sunTimes.Set.ToTime(nil).In(tz))
	fmt.Printf("Sun Culmination: %v | %v\n", sunTimes.Culmination.ToTime(nil).In(tz), sunTimes.LowerCulmination.ToTime(nil).In(tz))
	fmt.Printf("Sun Civil : %v | %v\n", sunTimes.Civil.Dawn.ToTime(nil).In(tz), sunTimes.Civil.Dusk.ToTime(nil).In(tz))

	// 月亮
	moonTimes, err := astronomy.MoonTwilight(jd, geo, false)
	if err != nil {
		fmt.Printf("MoonTwilight Error: %s", err.Error())
	}
	fmt.Printf("Moon Rise: %v\n", moonTimes.Rise.ToTime(nil).In(tz))
	fmt.Printf("Moon Set: %v\n", moonTimes.Set.ToTime(nil).In(tz))
	fmt.Printf("Moon Culmination: %v | %v\n", moonTimes.Culmination.ToTime(nil).In(tz), moonTimes.LowerCulmination.ToTime(nil).In(tz))

}
