package characters

import "maps"

type MobMasteries struct {
	Tame map[int]int `json:"tame,omitempty"` // mobId to proficiency
}

func (m MobMasteries) Clone() MobMasteries {
	m.Tame = maps.Clone(m.Tame)
	return m
}

// // // // // // // // // // // //
// Tame related
// // // // // // // // // // // //

func (m *MobMasteries) GetAllTame() map[int]int {

	retMap := map[int]int{}

	if m.Tame == nil {
		return retMap
	}

	for k, v := range m.Tame {
		retMap[k] = v
	}

	return retMap
}

func (m *MobMasteries) GetTame(mobId int) int {

	if m.Tame == nil {
		return 0
	}

	return m.Tame[mobId]
}

func (m *MobMasteries) SetTame(mobId int, amt int) {

	if m.Tame == nil {
		m.Tame = map[int]int{}
	}

	m.Tame[mobId] = amt
}
