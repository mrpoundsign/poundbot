package raid

import (
	"github.com/poundbot/poundbot/pkg/models"
)

type testRaidHandler struct {
	RaidAlert *models.RaidAlert
}

func (rh *testRaidHandler) RaidNotify(ra models.RaidAlertWithMessageChannel) {
	rh.RaidAlert = &ra.RaidAlert
}

// func TestAlerter_Run(t *testing.T) {
// 	t.Parallel()

// 	miu := func(ra models.RaidAlertWithMessageChannel, is messageIDSetter) {}

// 	var ra = models.RaidAlert{PlayerID: "1234"}

// 	tests := []struct {
// 		name       string
// 		raidAlerts []models.RaidAlert
// 		want       *models.RaidAlert
// 	}{
// 		{
// 			name: "With nothing",
// 		},
// 		{
// 			name:       "With RaidAlert",
// 			raidAlerts: []models.RaidAlert{ra},
// 			want:       &ra,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// var hit bool
// 			done := make(chan interface{}, 1)

// 			mockRH := &testRaidHandler{}

// 			mockRA := mocks.RaidAlertsStore{}

// 			mockRA.On("GetReady").
// 				Return(func() []models.RaidAlert {
// 					done <- nil
// 					return tt.raidAlerts
// 				}, nil)

// 			if len(tt.raidAlerts) != 0 {
// 				mockRA.On("IncrementNotifyCount", ra).Return(nil).Once()
// 				mockRA.On("Remove", ra).Return(nil).Once()
// 			}

// 			raidAlerter := NewAlerter(&mockRA, mockRH, done)
// 			raidAlerter.SleepTime = 1 * time.Microsecond
// 			raidAlerter.miu = miu
// 			raidAlerter.Run()
// 			mockRA.AssertExpectations(t)
// 			assert.EqualValues(t, tt.want, mockRH.RaidAlert, "They should be equal")
// 		})
// 	}
// }
