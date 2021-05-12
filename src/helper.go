package main

import (
	"fmt"
	"time"
)

// validate birthday
func validateAge(birthdate time.Time, checkage int) bool {
	today := time.Now()
	today = today.In(birthdate.Location())
	ty, tm, td := today.Date()
	today = time.Date(ty, tm, td, 0, 0, 0, 0, time.UTC)
	by, bm, bd := birthdate.Date()
	birthdate = time.Date(by, bm, bd, 0, 0, 0, 0, time.UTC)
	if today.Before(birthdate) {
		return false
	}
	age := ty - by
	anniversary := birthdate.AddDate(age, 0, 0)
	if anniversary.After(today) {
		age--
	}

	if age < checkage {
		fmt.Println("Ist noch NICHT 18 Jahre!!!")
		return false
	}
	return true
}
