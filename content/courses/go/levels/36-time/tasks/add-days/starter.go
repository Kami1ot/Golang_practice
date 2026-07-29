package main

import (
	"fmt"
	"time"
)

const layout = "2006-01-02"

func main() {
	var date string
	var days int
	fmt.Scan(&date, &days)

	// TODO: разберите date через time.Parse(layout, ...),
	// прибавьте days дней через AddDate
	// и напечатайте результат в формате layout
	_, _ = time.Parse(layout, date)
	_ = days
}
