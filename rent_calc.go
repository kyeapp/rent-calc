package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"
)

// TODO: change this to ask the user
var s = "Oct 20, 2018"
var e = "Nov 19, 2018"
var billType = "PG&E"
var billAmount = 72.28

type tennant struct {
	id          int
	name        string
	email       string
	phoneNumber string
	moveInDate  time.Time
	moveOutDate time.Time // Will be January 1, year 1, 00:00:00 UTC is tennant has not moved out
	room        string
	roommate    int // no roommate has a value of -1
}

// Return true is the tennant currently lives in the house
func (t *tennant) isCurrent() bool {
	return t.moveOutDate.IsZero()
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// returns a map of all tennants ID to their information
func loadTennants() (map[int]tennant, error) {
	tennant_file, err := os.Open("./resources/tennants.csv")
	checkErr(err)
	defer tennant_file.Close()

	csvReader := csv.NewReader(tennant_file)
	tennantMap := make(map[int]tennant, 0)

	// Discard the first line, which should contain headers
	_, _ = csvReader.Read()

	// read in tennants
	for {
		v, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		checkErr(err)

		// get tennant ID, and roommate if applicable
		t_id, err := strconv.Atoi(v[0])
		checkErr(err)

		t_roommate := -1
		if v[7] != "" {
			t_roommate, err = strconv.Atoi(v[7])
			checkErr(err)
		}

		// parse tennant move in and move out dates
		dateFormat := "Jan _2, 2006"

		var t_moveInDate time.Time
		t_moveInDate, err = time.Parse(dateFormat, v[4])
		checkErr(err)

		var t_moveOutDate time.Time
		if v[5] != "" {
			t_moveOutDate, err = time.Parse(dateFormat, v[5])
			checkErr(err)
		}

		t := tennant{
			id:          t_id,
			name:        v[1],
			email:       v[2],
			phoneNumber: v[3],
			moveInDate:  t_moveInDate,
			moveOutDate: t_moveOutDate,
			room:        v[6],
			roommate:    t_roommate,
		}
		tennantMap[t_id] = t
	}

	return tennantMap, nil
}

type billableDays struct {
	startDate time.Time
	endDate   time.Time
	total     int
}

// Returns
// 1. a map of tennant ID to the number of days they can be billed for
// 2. a total number of tennant days
func findBillableTennantDays(tennants map[int]tennant, billPeriodStart time.Time, billPeriodEnd time.Time) (map[int]billableDays, int) {
	billableDaysMap := make(map[int]billableDays, 0)
	var totalTennantDays int

	for _, tennant := range tennants {
		// skip tennants that have moved out before the billing period starts
		if !tennant.isCurrent() && tennant.moveOutDate.Before(billPeriodStart) {
			continue
		}

		// Find date when the tennant should be started billing
		tennantBillStartDate := billPeriodStart
		if billPeriodStart.Before(tennant.moveInDate) {
			tennantBillStartDate = tennant.moveInDate
		}

		// Find date when the tennant should end billing
		tennantBillEndDate := billPeriodEnd
		if !tennant.isCurrent() && tennant.moveOutDate.Before(billPeriodEnd) {
			tennantBillEndDate = tennant.moveOutDate
		}

		// Calculate how many days tennant should be billed for
		tennantTotal := int(tennantBillEndDate.Sub(tennantBillStartDate).Hours()/24) + 1 // add one because we want inclusive end date

		billableDaysMap[tennant.id] = billableDays{
			startDate: tennantBillStartDate,
			endDate:   tennantBillEndDate,
			total:     tennantTotal,
		}

		totalTennantDays += tennantTotal

		fmt.Println("TENANT INFO===================================")
		fmt.Println(tennant.name, tennantBillStartDate.Format("Jan 2, 2006"), tennantBillEndDate.Format("Jan 2, 2006"), tennantTotal)
		fmt.Println()
	}

	return billableDaysMap, totalTennantDays
}

func main() {
	AllTennantsMap, err := loadTennants()
	checkErr(err)

	// Take input for the billing dates and the amount that is due
	dateFormat := "Jan _2, 2006"

	billPeriodStart, err := time.Parse(dateFormat, s)
	checkErr(err)

	billPeriodEnd, err := time.Parse(dateFormat, e)
	checkErr(err)

	fmt.Println(" before billable function ========================")
	fmt.Println("start period:", billPeriodStart.Format("Jan 2, 2006"))
	fmt.Println("end period:", billPeriodEnd.Format("Jan 2, 2006"))

	// find what tennats can be billed for that timeframe and for how many days
	billableDaysMap, totalTennantDays := findBillableTennantDays(AllTennantsMap, billPeriodStart, billPeriodEnd)
	checkErr(err)

	fmt.Println("========================")
	for _, t := range billableDaysMap {
		fmt.Println(t)
	}

	//do the calulation

	//show how much each person owes for that bill.

	fmt.Println(billType)
	fmt.Println(billAmount)
	fmt.Println(totalTennantDays)

}
