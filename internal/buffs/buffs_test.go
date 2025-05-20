package buffs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDurations(t *testing.T) {
	type args struct {
		buff *Buff
		spec *BuffSpec
	}
	tests := []struct {
		name       string
		args       args
		wantRounds int
		wantTotal  int
	}{
		{
			name: "Normal case",
			args: args{
				buff: &Buff{RoundCounter: 2},
				spec: &BuffSpec{TriggerCount: 5, RoundInterval: 3},
			},
			wantRounds: 13, // (5*3)-2 = 15-2 = 13
			wantTotal:  15,
		},
		{
			name: "Zero rounds passed",
			args: args{
				buff: &Buff{RoundCounter: 0},
				spec: &BuffSpec{TriggerCount: 4, RoundInterval: 2},
			},
			wantRounds: 8,
			wantTotal:  8,
		},
		{
			name: "All rounds passed",
			args: args{
				buff: &Buff{RoundCounter: 12},
				spec: &BuffSpec{TriggerCount: 3, RoundInterval: 4},
			},
			wantRounds: 0, // (3*4)-12 = 12-12 = 0
			wantTotal:  12,
		},
		{
			name: "RoundCounter greater than total",
			args: args{
				buff: &Buff{RoundCounter: 10},
				spec: &BuffSpec{TriggerCount: 2, RoundInterval: 4},
			},
			wantRounds: -2, // (2*4)-10 = 8-10 = -2
			wantTotal:  8,
		},
		{
			name: "Zero trigger count",
			args: args{
				buff: &Buff{RoundCounter: 1},
				spec: &BuffSpec{TriggerCount: 0, RoundInterval: 5},
			},
			wantRounds: -1, // (0*5)-1 = 0-1 = -1
			wantTotal:  0,
		},
		{
			name: "Zero round interval",
			args: args{
				buff: &Buff{RoundCounter: 3},
				spec: &BuffSpec{TriggerCount: 4, RoundInterval: 0},
			},
			wantRounds: -3, // (4*0)-3 = 0-3 = -3
			wantTotal:  0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotRounds, gotTotal := GetDurations(tt.args.buff, tt.args.spec)
			assert.Equal(t, tt.wantRounds, gotRounds)
			assert.Equal(t, tt.wantTotal, gotTotal)
		})
	}
}
func TestBuffs_HasBuff(t *testing.T) {
	type fields struct {
		list    []*Buff
		buffIds map[int]int
	}
	tests := []struct {
		name   string
		fields fields
		arg    int
		want   bool
	}{
		{
			name: "Buff exists in buffIds",
			fields: fields{
				list: []*Buff{
					{BuffId: 1},
					{BuffId: 2},
				},
				buffIds: map[int]int{1: 0, 2: 1},
			},
			arg:  1,
			want: true,
		},
		{
			name: "Buff does not exist in buffIds",
			fields: fields{
				list: []*Buff{
					{BuffId: 1},
					{BuffId: 2},
				},
				buffIds: map[int]int{1: 0, 2: 1},
			},
			arg:  3,
			want: false,
		},
		{
			name: "Empty buffIds map",
			fields: fields{
				list:    []*Buff{},
				buffIds: map[int]int{},
			},
			arg:  1,
			want: false,
		},
		{
			name: "BuffIds map is nil",
			fields: fields{
				list:    []*Buff{},
				buffIds: nil,
			},
			arg:  1,
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bs := &Buffs{
				List:    tt.fields.list,
				buffIds: tt.fields.buffIds,
			}
			assert.Equal(t, tt.want, bs.HasBuff(tt.arg))
		})
	}
}
func TestBuffs_Started(t *testing.T) {
	type fields struct {
		list    []*Buff
		buffIds map[int]int
	}
	tests := []struct {
		name         string
		fields       fields
		arg          int
		wantOnStart  bool
		shouldChange bool
	}{
		{
			name: "Buff exists and OnStartWaiting is true",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, OnStartWaiting: true},
					{BuffId: 2, OnStartWaiting: true},
				},
				buffIds: map[int]int{1: 0, 2: 1},
			},
			arg:          1,
			wantOnStart:  false,
			shouldChange: true,
		},
		{
			name: "Buff exists and OnStartWaiting is already false",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, OnStartWaiting: false},
				},
				buffIds: map[int]int{1: 0},
			},
			arg:          1,
			wantOnStart:  false,
			shouldChange: false,
		},
		{
			name: "Buff does not exist",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, OnStartWaiting: true},
				},
				buffIds: map[int]int{1: 0},
			},
			arg:          2,
			wantOnStart:  true,
			shouldChange: false,
		},
		{
			name: "Empty buffIds map",
			fields: fields{
				list:    []*Buff{},
				buffIds: map[int]int{},
			},
			arg:          1,
			wantOnStart:  false,
			shouldChange: false,
		},
		{
			name: "buffIds is nil",
			fields: fields{
				list:    []*Buff{},
				buffIds: nil,
			},
			arg:          1,
			wantOnStart:  false,
			shouldChange: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bs := &Buffs{
				List:    tt.fields.list,
				buffIds: tt.fields.buffIds,
			}
			if tt.shouldChange && len(bs.List) > 0 {
				assert.True(t, bs.List[bs.buffIds[tt.arg]].OnStartWaiting)
			}
			bs.Started(tt.arg)
			if idx, ok := bs.buffIds[tt.arg]; ok && idx < len(bs.List) {
				assert.Equal(t, tt.wantOnStart, bs.List[idx].OnStartWaiting)
			} else if len(bs.List) > 0 {
				// If buff does not exist, original value should remain unchanged
				assert.Equal(t, tt.wantOnStart, bs.List[0].OnStartWaiting)
			}
		})
	}
}
func TestBuffs_TriggersLeft(t *testing.T) {
	type fields struct {
		list    []*Buff
		buffIds map[int]int
	}
	tests := []struct {
		name   string
		fields fields
		arg    int
		want   int
	}{
		{
			name: "Buff exists and has positive TriggersLeft",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, TriggersLeft: 3},
					{BuffId: 2, TriggersLeft: 5},
				},
				buffIds: map[int]int{1: 0, 2: 1},
			},
			arg:  2,
			want: 5,
		},
		{
			name: "Buff exists and has zero TriggersLeft",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, TriggersLeft: 0},
				},
				buffIds: map[int]int{1: 0},
			},
			arg:  1,
			want: 0,
		},
		{
			name: "Buff exists and has negative TriggersLeft",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, TriggersLeft: -2},
				},
				buffIds: map[int]int{1: 0},
			},
			arg:  1,
			want: -2,
		},
		{
			name: "Buff does not exist in buffIds",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, TriggersLeft: 3},
				},
				buffIds: map[int]int{1: 0},
			},
			arg:  2,
			want: 0,
		},
		{
			name: "buffIds is nil",
			fields: fields{
				list:    []*Buff{{BuffId: 1, TriggersLeft: 7}},
				buffIds: nil,
			},
			arg:  1,
			want: 0,
		},
		{
			name: "buffIds is empty map",
			fields: fields{
				list:    []*Buff{{BuffId: 1, TriggersLeft: 7}},
				buffIds: map[int]int{},
			},
			arg:  1,
			want: 0,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bs := &Buffs{
				List:    tt.fields.list,
				buffIds: tt.fields.buffIds,
			}
			assert.Equal(t, tt.want, bs.TriggersLeft(tt.arg))
		})
	}
}
func TestBuffs_RemoveBuff(t *testing.T) {
	type fields struct {
		list    []*Buff
		buffIds map[int]int
	}
	tests := []struct {
		name         string
		fields       fields
		arg          int
		wantResult   bool
		wantTriggers int
		shouldModify bool
	}{
		{
			name: "Buff exists and is removed",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, TriggersLeft: 5},
					{BuffId: 2, TriggersLeft: 3},
				},
				buffIds: map[int]int{1: 0, 2: 1},
			},
			arg:          2,
			wantResult:   true,
			wantTriggers: TriggersLeftExpired,
			shouldModify: true,
		},
		{
			name: "Buff does not exist",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, TriggersLeft: 5},
				},
				buffIds: map[int]int{1: 0},
			},
			arg:          3,
			wantResult:   false,
			wantTriggers: 5,
			shouldModify: false,
		},
		{
			name: "buffIds is nil",
			fields: fields{
				list:    []*Buff{{BuffId: 1, TriggersLeft: 7}},
				buffIds: nil,
			},
			arg:          1,
			wantResult:   false,
			wantTriggers: 7,
			shouldModify: false,
		},
		{
			name: "buffIds is empty map",
			fields: fields{
				list:    []*Buff{{BuffId: 1, TriggersLeft: 7}},
				buffIds: map[int]int{},
			},
			arg:          1,
			wantResult:   false,
			wantTriggers: 7,
			shouldModify: false,
		},
		{
			name: "Multiple buffs, remove first",
			fields: fields{
				list: []*Buff{
					{BuffId: 10, TriggersLeft: 2},
					{BuffId: 20, TriggersLeft: 4},
				},
				buffIds: map[int]int{10: 0, 20: 1},
			},
			arg:          10,
			wantResult:   true,
			wantTriggers: TriggersLeftExpired,
			shouldModify: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bs := &Buffs{
				List:    tt.fields.list,
				buffIds: tt.fields.buffIds,
			}
			result := bs.RemoveBuff(tt.arg)
			assert.Equal(t, tt.wantResult, result)
			if idx, ok := bs.buffIds[tt.arg]; ok && tt.shouldModify {
				assert.Equal(t, tt.wantTriggers, bs.List[idx].TriggersLeft)
			} else if len(bs.List) > 0 {
				// If not modified, original value should remain
				assert.Equal(t, tt.wantTriggers, bs.List[0].TriggersLeft)
			}
		})
	}
}
func TestBuff_Expired(t *testing.T) {
	tests := []struct {
		name         string
		triggersLeft int
		want         bool
	}{
		{
			name:         "TriggersLeft is zero (expired)",
			triggersLeft: 0,
			want:         true,
		},
		{
			name:         "TriggersLeft is negative (expired)",
			triggersLeft: -1,
			want:         true,
		},
		{
			name:         "TriggersLeft is positive (not expired)",
			triggersLeft: 3,
			want:         false,
		},
		{
			name:         "TriggersLeft is TriggersLeftUnlimited (not expired)",
			triggersLeft: TriggersLeftUnlimited,
			want:         false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Buff{TriggersLeft: tt.triggersLeft}
			assert.Equal(t, tt.want, b.Expired())
		})
	}
}
