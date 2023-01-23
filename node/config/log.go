package config

type Log struct {
	IncludeEvents []string       `yaml:"include_events"`
	ExcludeEvents []string       `yaml:"exclude_events"`
	TagLevels     map[string]int `yaml:"tag_levels"`
	HideDate      bool           `yaml:"hide_date"`
}

func (l *Log) isIncluded(event string) bool {
	if len(l.IncludeEvents) == 0 {
		return false
	}
	for _, e := range l.IncludeEvents {
		if e == event {
			return true
		}
	}
	return false
}

func (l *Log) isExcluded(event string) bool {
	if len(l.ExcludeEvents) == 0 {
		return false
	}
	for _, e := range l.ExcludeEvents {
		if e == event {
			return true
		}
	}
	return false
}

func (l *Log) IsEventLoggable(event string) bool {
	if len(l.ExcludeEvents) > 0 {
		return !l.isExcluded(event)
	}

	return l.isIncluded(event)
}
