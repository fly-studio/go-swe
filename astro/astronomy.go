package astro

import (
	"go-swe/swe"
	"math"
)

type Astronomy struct {
	// 地理位置
	Geo GeographicCoordinates
	// 儒略日(世界时)
	JdUt JulianDay
	// 儒略日(天文时)
	JdEt JulianDay
	// Swe的实例
	Swe swe.SweInterface
}

type EclipticProperties struct {
	// 真黄道真倾角(含章动)，即黄赤交角，转轴倾角
	// true obliquity of the Ecliptic (includes nutation)
	TrueObliquity float64
	// 黄道平均倾角
	// 比如地球是：23°26′20.512″
	// mean obliquity of the Ecliptic
	MeanObliquity float64
	// 黄经章动
	// nutation in longitude
	NutationInLongitude float64
	// 倾角章动
	// nutation in obliquity
	NutationInObliquity float64
}

type PlanetProperties struct {
	// 黄道坐标
	Ecliptic EclipticCoordinates
	// 距离
	Distance float64
	// 经度的速度
	SpeedInLongitude float64
	// 纬度的速度
	SpeedInLatitude float64
	// 距离的速度
	SpeedInDistance float64
}

func NewAstronomy(longitude, latitude float64, jdUt JulianDay) *Astronomy {

	_swe := swe.NewSwe()
	return &Astronomy{
		Geo: GeographicCoordinates{
			Longitude: longitude,
			Latitude: latitude,
		},
		JdUt: jdUt,
		JdEt: jdUt.ToEphemerisTime(JulianDayDelta(_swe.DeltaT(float64(jdUt)))),
		Swe:  _swe,
	}
}

func (astro *Astronomy) simpleCalcFlags() *swe.CalcFlags {
	var iFlag int32 = swe.FlagEphSwiss | swe.FlagRadians | swe.FlagSpeed
	return &swe.CalcFlags{
		Flags: iFlag,
	}
}

/**
 * 黄道倾角
 */
func (astro *Astronomy) Ecliptic() (*EclipticProperties, error) {

	// 黄道章动
	res, _, err := astro.Swe.Calc(float64(astro.JdEt), swe.EclNut, astro.simpleCalcFlags())

	if err != nil {
		return nil, err
	}

	return &EclipticProperties{
		TrueObliquity:       res[0],
		MeanObliquity:       res[1],
		NutationInLongitude: res[2],
		NutationInObliquity: res[3],
	}, nil
}

/**
 * 星星当前的属性，包括黄道经纬，距离，黄道经纬速度，具体速度
 */
func (astro *Astronomy) Planet(planetId swe.Planet) (*PlanetProperties, error) {
	// 黄道章动
	res, _, err := astro.Swe.Calc(float64(astro.JdEt), planetId, astro.simpleCalcFlags())

	if err != nil {
		return nil, err
	}

	return &PlanetProperties{
		Ecliptic: EclipticCoordinates{
			Longitude: res[0],
			Latitude: res[1],
		},
		Distance:    0,
		SpeedInLongitude: res[3],
		SpeedInLatitude: res[4],
		SpeedInDistance: res[5],
	}, nil
}

/**
 * 星星当前时刻的时角, 星星属性,
 * 章动同时影响恒星时和天体坐标,所以不计算章动。
 * withRevise 是否修正，包含使用真黄道倾角、修正大气折射、修正地平坐标中视差
 */
func (astro *Astronomy) PlanetHourAngle(planetId swe.Planet, withRevise bool) (hourAngle float64, planet *PlanetProperties, equatorial *EquatorialCoordinates, err error) {
	// 当前黄道倾角、章动等参数
	ecliptic, err := astro.Ecliptic()
	if err != nil {
		return
	}

	// 星星的黄道等参数
	planet, err = astro.Planet(planetId)
	if err != nil {
		return
	}

	// 修正光行差 20.5″
	planet.Ecliptic.Longitude -= 20.5 / DegreeSecondsPerRadian

    // 黄道坐标 -> 赤道坐标
	equatorial = EclipticToEquatorial(&planet.Ecliptic, IfThenElse(withRevise, ecliptic.TrueObliquity, ecliptic.MeanObliquity).(float64))

	// 不太精确的恒星时
	sidTime := GreenwichMeridianSiderealTime(astro.JdEt)
	// 修正恒星时
	if withRevise {
		sidTime += ecliptic.NutationInLongitude * math.Cos(ecliptic.TrueObliquity)
	}

	/*如果θ是本地恒星时，θo是格林尼治恒星时，L是观者站经度（从格林尼治向西为正，东为负），α是赤道经度，那么本地时角H计算如下：
	H = θ - α 或 H =θo - L - α
	如果α含章动效果，那么H也含章动（见11章）。*/

	// 时角
	hourAngle = RadianMod(sidTime - astro.Geo.Longitude - equatorial.RightAscension)
	if hourAngle > Radian180 {
		// 得到此刻天体时角
		hourAngle -= Radian360
	}

	return
}
