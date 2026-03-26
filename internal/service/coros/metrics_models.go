package coros

type AccountQueryResponse struct {
	Result string           `json:"result"`
	Msg    string           `json:"message"`
	Data   AccountQueryData `json:"data"`
}

type AccountQueryData struct {
	Weight      float64         `json:"weight"`
	Stature     float64         `json:"stature"`
	Birthday    int             `json:"birthday"`
	Sex         int             `json:"sex"`
	CountryCode string          `json:"countryCode"`
	MaxHr       float64         `json:"maxHr"`
	Rhr         float64         `json:"rhr"`
	HrZoneType  int             `json:"hrZoneType"`
	ZoneData    AccountZoneData `json:"zoneData"`
}

type AccountZoneData struct {
	LTHR     float64            `json:"lthr"`
	LTSP     float64            `json:"ltsp"`
	LTHRZone []ZoneBoundary     `json:"lthrZone"`
	LTSPZone []PaceZoneBoundary `json:"ltspZone"`
}

type ZoneBoundary struct {
	Hr    float64 `json:"hr"`
	Ratio float64 `json:"ratio"`
}

type PaceZoneBoundary struct {
	Pace  float64 `json:"pace"`
	Ratio float64 `json:"ratio"`
}

type DashboardQueryResponse struct {
	Result string             `json:"result"`
	Msg    string             `json:"message"`
	Data   DashboardQueryData `json:"data"`
}

type DashboardQueryData struct {
	SummaryInfo DashboardSummaryInfo `json:"summaryInfo"`
}

type DashboardSummaryInfo struct {
	StaminaLevel                  float64            `json:"staminaLevel"`
	AerobicEnduranceScore         float64            `json:"aerobicEnduranceScore"`
	LactateThresholdCapacityScore float64            `json:"lactateThresholdCapacityScore"`
	AnaerobicEnduranceScore       float64            `json:"anaerobicEnduranceScore"`
	AnaerobicCapacityScore        float64            `json:"anaerobicCapacityScore"`
	StaminaLevelRanking           float64            `json:"staminaLevelRanking"`
	LTHR                          float64            `json:"lthr"`
	LTSP                          float64            `json:"ltsp"`
	FitnessMaxHr                  float64            `json:"fitnessMaxHr"`
	Rhr                           float64            `json:"rhr"`
	LTHRZone                      []ZoneBoundary     `json:"lthrZone"`
	LTSPZone                      []PaceZoneBoundary `json:"ltspZone"`
	RecoveryPct                   float64            `json:"recoveryPct"`
	RecoveryState                 int                `json:"recoveryState"`
	FullRecoveryHours             float64            `json:"fullRecoveryHours"`
	SleepHrvData                  DashboardHRVData   `json:"sleepHrvData"`
	RecordDetailList              []RecordDetail     `json:"recordDetailList"`
}

type DashboardHRVData struct {
	AvgSleepHRV  float64         `json:"avgSleepHrv"`
	SleepHrvBase float64         `json:"sleepHrvBase"`
	SleepHrvSD   float64         `json:"sleepHrvSd"`
	SleepHrvList []HRVDailyValue `json:"sleepHrvList"`
}

type HRVDailyValue struct {
	HappenDay int     `json:"happenDay"`
	Value     float64 `json:"value"`
}

type RecordDetail struct {
	Type   int     `json:"type"`
	Record float64 `json:"record"`
}

type DashboardDetailQueryResponse struct {
	Result string                   `json:"result"`
	Msg    string                   `json:"message"`
	Data   DashboardDetailQueryData `json:"data"`
}

type DashboardDetailQueryData struct {
	SummaryInfo       DashboardDetailSummaryInfo `json:"summaryInfo"`
	DetailList        []TrainingTrendPoint       `json:"detailList"`
	SportDataList     []RecentSportData          `json:"sportDataList"`
	CurrentWeekRecord WeekRecord                 `json:"currentWeekRecord"`
}

type DashboardDetailSummaryInfo struct {
	ATI                        float64 `json:"ati"`
	CTI                        float64 `json:"cti"`
	TrainingLoadRatio          float64 `json:"trainingLoadRatio"`
	TrainingLoadRatioState     int     `json:"trainingLoadRatioState"`
	TiredRateNewPercentInState float64 `json:"tiredRateNewPercentInState"`
	TiredRate                  float64 `json:"tiredRate"`
	TiredRateNew               float64 `json:"tiredRateNew"`
	TiredRateNewState          int     `json:"tiredRateNewState"`
	RecomendTlInDays           float64 `json:"recomendTlInDays"`
}

type TrainingTrendPoint struct {
	HappenDay         int     `json:"happenDay"`
	TrainingLoad      float64 `json:"trainingLoad"`
	ATI               float64 `json:"ati"`
	CTI               float64 `json:"cti"`
	TiredRate         float64 `json:"tiredRate"`
	TiredRateNew      float64 `json:"tiredRateNew"`
	TiredRateStateNew int     `json:"tiredRateStateNew"`
	VO2Max            float64 `json:"vo2max"`
	LTHR              float64 `json:"lthr"`
	LTSP              float64 `json:"ltsp"`
	StaminaLevel      float64 `json:"staminaLevel"`
	Performance       float64 `json:"performance"`
}

type RecentSportData struct {
	HappenDay      int     `json:"happenDay"`
	LabelID        string  `json:"labelId"`
	SportType      int     `json:"sportType"`
	Mode           int     `json:"mode"`
	SubMode        int     `json:"subMode"`
	Distance       float64 `json:"distance"`
	Duration       float64 `json:"duration"`
	AvgPace        float64 `json:"avgPace"`
	AvgHeartRate   float64 `json:"avgHeartRate"`
	AvgPower       float64 `json:"avgPower"`
	Step           float64 `json:"step"`
	TrainingLoad   float64 `json:"trainingLoad"`
	TotalElevation float64 `json:"totalElevation"`
}

type WeekRecord struct {
	DistanceRecord AggregateRecord `json:"distanceRecord"`
	DurationRecord AggregateRecord `json:"durationRecord"`
	TLRecord       AggregateRecord `json:"tlRecord"`
}

type AggregateRecord struct {
	Count       int                    `json:"count"`
	DetailList  []AggregateRecordPoint `json:"detailList"`
	Percentage  float64                `json:"percentage"`
	TotalTarget float64                `json:"totalTarget"`
	TotalValue  float64                `json:"totalValue"`
	Type        int                    `json:"type"`
}

type AggregateRecordPoint struct {
	Count             int     `json:"count"`
	HappenDay         int     `json:"happenDay"`
	PeriodHighPct     float64 `json:"periodHighPct"`
	PeriodHighValue   float64 `json:"periodHighValue"`
	PeriodLowPct      float64 `json:"periodLowPct"`
	PeriodLowValue    float64 `json:"periodLowValue"`
	PeriodMediumPct   float64 `json:"periodMediumPct"`
	PeriodMediumValue float64 `json:"periodMediumValue"`
	Target            float64 `json:"target"`
	Timestamp         int64   `json:"timestamp"`
	Value             float64 `json:"value"`
}
