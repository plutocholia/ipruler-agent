package config

type SettingsModel struct {
	TableHardSync []int `yaml:"table-hard-sync"`
}

func (s *SettingsModel) IsEmpty() bool {
	if len(s.TableHardSync) == 0 {
		return true
	}
	return false
}
