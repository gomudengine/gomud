package configs

type ProgressionConfig struct {
	// Stat gain formula: GainsForLevel
	//   racial_value = floor(base * BaseModFactor * (level-1)^BaseModExponent)
	//                + floor(NaturalGainsModFactor * level^NaturalGainsExponent)
	//
	// BaseModFactor controls how much a racial base stat matters at high levels.
	// Higher values cause races with strong bases to diverge more from weak-base races.
	BaseModFactor ConfigFloat `yaml:"BaseModFactor"`
	// BaseModExponent controls the shape of the base-scaled component.
	// 1.0 = linear growth, <1.0 = diminishing returns, >1.0 = accelerating returns.
	BaseModExponent ConfigFloat `yaml:"BaseModExponent"`
	// NaturalGainsModFactor controls the universal flat gains every character receives
	// per level, regardless of race. Higher values raise the floor for weak-base races.
	NaturalGainsModFactor ConfigFloat `yaml:"NaturalGainsModFactor"`
	// NaturalGainsExponent controls the shape of the flat gains component.
	// 1.0 = linear, <1.0 = diminishing returns, >1.0 = accelerating.
	NaturalGainsExponent ConfigFloat `yaml:"NaturalGainsExponent"`

	// HP formula: HealthMax = HPBase + level*HPPerLevel + Vitality_adj*HPPerVitality + mods
	HPBase        ConfigInt   `yaml:"HPBase"`
	HPPerLevel    ConfigFloat `yaml:"HPPerLevel"`
	HPPerVitality ConfigFloat `yaml:"HPPerVitality"`

	// Mana formula: ManaMax = ManaBase + level*ManaPerLevel + Mysticism_adj*ManaPerMysticism + mods
	ManaBase         ConfigInt   `yaml:"ManaBase"`
	ManaPerLevel     ConfigFloat `yaml:"ManaPerLevel"`
	ManaPerMysticism ConfigFloat `yaml:"ManaPerMysticism"`

	// Points awarded to the player on each level-up.
	TrainingPointsPerLevel     ConfigInt `yaml:"TrainingPointsPerLevel"`
	TrainingPointsEveryNLevels ConfigInt `yaml:"TrainingPointsEveryNLevels"`
	StatPointsPerLevel         ConfigInt `yaml:"StatPointsPerLevel"`
	StatPointsEveryNLevels     ConfigInt `yaml:"StatPointsEveryNLevels"`

	// XP curve: XP_to_level(L) = (XPBase + L^XPLevelPower * XPLevelFactor * XPBase) * TNLScale
	// XPBase is the flat XP cost at level 1.
	XPBase ConfigInt `yaml:"XPBase"`
	// XPLevelFactor scales how fast the curve rises. Higher = more XP per level.
	XPLevelFactor ConfigFloat `yaml:"XPLevelFactor"`
	// XPLevelPower is the exponent. 2.0 = quadratic, 1.5 = gentler, 3.0 = steeper.
	XPLevelPower ConfigFloat `yaml:"XPLevelPower"`

	// MaxLevel is the soft display cap used in admin charts and any future level-cap enforcement.
	MaxLevel ConfigInt `yaml:"MaxLevel"`

	// Stat value compression (applied in Recalculate).
	// Once a stat's Value reaches StatCapThreshold, further gains are compressed:
	//   ValueAdj = StatCapAnchor + round((Value-StatCapAnchor)^StatCapExponent * StatCapScale)
	// StatCapThreshold: the Value at which compression begins. Default 105.
	StatCapThreshold ConfigInt `yaml:"StatCapThreshold"`
	// StatCapAnchor: the value the compression formula is anchored to. Default 100.
	// Overage is measured from this point: overage = Value - StatCapAnchor.
	StatCapAnchor ConfigInt `yaml:"StatCapAnchor"`
	// StatCapExponent: controls how aggressively gains are compressed above the threshold.
	// 0.5 = sqrt (default, strong compression), 1.0 = linear pass-through (no compression),
	// 0.25 = very aggressive. Valid range: 0.01 to 1.0.
	StatCapExponent ConfigFloat `yaml:"StatCapExponent"`
	// StatCapScale: multiplier applied after the exponent. Default 2.0.
	StatCapScale ConfigFloat `yaml:"StatCapScale"`
	// StatCapExemptBonus: when true, only the racial portion of a stat is compressed.
	// Training points and equipment/buff mods are added on top of the compressed racial
	// value without being subject to the cap. This ensures deliberate investment always
	// pays off fully. Default false (original behaviour: everything compressed together).
	StatCapExemptBonus ConfigBool `yaml:"StatCapExemptBonus"`
}

func (p *ProgressionConfig) Validate() {
	if p.BaseModFactor <= 0 {
		p.BaseModFactor = 0.3333333334
	}
	if p.BaseModExponent < 0.1 || p.BaseModExponent > 5.0 {
		p.BaseModExponent = 1.0
	}
	if p.NaturalGainsModFactor <= 0 {
		p.NaturalGainsModFactor = 0.5
	}
	if p.NaturalGainsExponent < 0.1 || p.NaturalGainsExponent > 5.0 {
		p.NaturalGainsExponent = 1.0
	}

	if p.HPBase < 0 {
		p.HPBase = 5
	}
	if p.HPPerLevel <= 0 {
		p.HPPerLevel = 1.0
	}
	if p.HPPerVitality <= 0 {
		p.HPPerVitality = 4.0
	}

	if p.ManaBase < 0 {
		p.ManaBase = 4
	}
	if p.ManaPerLevel <= 0 {
		p.ManaPerLevel = 1.0
	}
	if p.ManaPerMysticism <= 0 {
		p.ManaPerMysticism = 3.0
	}

	if p.TrainingPointsPerLevel < 0 {
		p.TrainingPointsPerLevel = 1
	}
	if p.TrainingPointsEveryNLevels < 1 {
		p.TrainingPointsEveryNLevels = 1
	}
	if p.StatPointsPerLevel < 0 {
		p.StatPointsPerLevel = 1
	}
	if p.StatPointsEveryNLevels < 1 {
		p.StatPointsEveryNLevels = 1
	}

	if p.XPBase < 1 {
		p.XPBase = 1000
	}
	if p.XPLevelFactor <= 0 {
		p.XPLevelFactor = 0.75
	}
	if p.XPLevelPower < 0.1 || p.XPLevelPower > 5.0 {
		p.XPLevelPower = 2.0
	}

	if p.MaxLevel < 10 {
		p.MaxLevel = 100
	}

	if p.StatCapThreshold < 1 {
		p.StatCapThreshold = 105
	}
	if p.StatCapAnchor < 0 {
		p.StatCapAnchor = 100
	}
	if p.StatCapAnchor >= p.StatCapThreshold {
		p.StatCapAnchor = p.StatCapThreshold - 1
	}
	if p.StatCapExponent <= 0 || p.StatCapExponent > 1.0 {
		p.StatCapExponent = 0.5
	}
	if p.StatCapScale <= 0 {
		p.StatCapScale = 2.0
	}
}

func GetProgressionConfig() ProgressionConfig {
	configDataLock.RLock()
	defer configDataLock.RUnlock()

	if !configData.validated {
		configData.Validate()
	}
	return configData.GamePlay.Progression
}
