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
const (
	DATE_FORMAT = "Jan _2, 2006"
	BILL_START  = "Oct 20, 2018"
	BILL_END    = "Nov 19, 2018"
	BILL_TYPE   = "PG&E"
	BILL_TOTAL  = 72.28
)

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

		var t_moveInDate time.Time
		t_moveInDate, err = time.Parse(DATE_FORMAT, v[4])
		checkErr(err)

		var t_moveOutDate time.Time
		if v[5] != "" {
			t_moveOutDate, err = time.Parse(DATE_FORMAT, v[5])
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

type bill struct {
	startDate    time.Time
	endDate      time.Time
	tennantDays  int
	tennantShare float64
	amountDue    float64
}

// Returns
// 1. a map of tennant ID to their bill (incomplete)
// 2. a total number of tennant days
func findBillableTennantDays(tennants map[int]tennant, billPeriodStart time.Time, billPeriodEnd time.Time) (map[int]*bill, int) {
	billMap := make(map[int]*bill, 0)
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
		tennantDays := int(tennantBillEndDate.Sub(tennantBillStartDate).Hours()/24) + 1 // add one because we want inclusive end date

		billMap[tennant.id] = &bill{
			startDate:   tennantBillStartDate,
			endDate:     tennantBillEndDate,
			tennantDays: tennantDays,
		}

		totalTennantDays += tennantDays
	}

	return billMap, totalTennantDays
}

// Verify that the calculations still add up
func verifyBillCalculations(billMap map[int]*bill) {
	var totalAmountDue float64
	var totalTennantShares float64
	for _, bill := range billMap {
		totalAmountDue += bill.amountDue
		totalTennantShares += bill.tennantShare
	}

	if totalAmountDue != BILL_TOTAL {
		panic("INDIVIDUAL BILLS DID NOT ADD UP TO TOTAL BILL")
	}
	if totalTennantShares != float64(1) {
		panic("TOTAL SHARES DO NOT ADD UP TO 100%")
	}
}

func main() {
	tennantsMap, err := loadTennants()
	checkErr(err)

	billPeriodStart, err := time.Parse(DATE_FORMAT, BILL_START)
	checkErr(err)

	billPeriodEnd, err := time.Parse(DATE_FORMAT, BILL_END)
	checkErr(err)

	// find what tennats can be billed for that timeframe and for how many days
	billMap, totalTennantDays := findBillableTennantDays(tennantsMap, billPeriodStart, billPeriodEnd)
	checkErr(err)

	//do the calulation
	for _, bill := range billMap {
		bill.tennantShare = float64(float64(bill.tennantDays) / float64(totalTennantDays))
		bill.amountDue = float64(BILL_TOTAL) * bill.tennantShare
	}

	verifyBillCalculations(billMap)

	//show how much each person owes for that bill.
	for tennantID, bill := range billMap {
		fmt.Println("Name:", tennantsMap[tennantID].name)
		fmt.Println("Bill type:", BILL_TYPE)
		fmt.Println("Start date (inclusive):", bill.startDate.Format(DATE_FORMAT))
		fmt.Println("End date: (inclusive)", bill.endDate.Format(DATE_FORMAT))
		fmt.Println("# of days:", bill.tennantDays)
		fmt.Printf("percent of bill: %.2f%% \n", bill.tennantShare*100)
		fmt.Printf("Amount Due: $%.2f \n", bill.amountDue)
		fmt.Println()
	}

}
