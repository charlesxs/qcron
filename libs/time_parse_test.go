package libs

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeParse(t *testing.T) {
	//cronStr := "* * * * * *"
	//cronStr := "*/2 * * * * *"
	//cronStr := "00 */10 * * * *"
	//cronStr := "00 */10 1,2,3 * * *"
	//cronStr := "00 1,2,3 */5 * * *"
	//cronStr := "00 12 */5 3-10 * *"
	cronStr := "00 12 */5 3,7,10 2,4 *"
	//cronStr := "00 12 */5 * 6,7 3-5"

	tt, err := TimeParse(cronStr)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("now:", time.Now())
	fmt.Println(tt.secondStart, tt.minuteStart, tt.hourStart, tt.dayStart, tt.monthStart)
	fmt.Println(tt.secondInterval, tt.minuteInterval, tt.hourInterval, tt.dayInterval, tt.monthInterval)
	fmt.Println(tt.Second, tt.Minute, tt.Hour, tt.Day, tt.Month, tt.Week)
	fmt.Println(tt.NextExecTime)
	for i := 0; i < 500; i++ {
		tt.ComputeNextExecTime(tt.NextExecTime)
		fmt.Println(tt.NextExecTime)
	}
}