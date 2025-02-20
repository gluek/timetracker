package handlers

import (
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	"github.com/gluek/timetracker/internal/database"

	"github.com/rickar/cal/v2"
	"github.com/spf13/viper"
)

func atoi(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		log.Println(err)
	}
	return value
}

func parseDate(date string) (int, int, int) {

	dateTime, _ := time.Parse("2006-01-02", date)
	year, month, day := dateTime.Date()
	monthInt := int(month)
	return year, monthInt, day
}

func findIDTimeframe(id int) int {
	for index, timeframe := range tfList {
		if timeframe.ID == id {
			return index
		}
	}
	return -1
}

func timeframeDuration(tf database.Timeframe) (time.Duration, error) {
	timeStart, err := time.Parse("15:04", tf.Start)
	if err != nil {
		return -1, err
	}
	timeEnd, err := time.Parse("15:04", tf.End)
	if err != nil {
		return -1, err
	}
	return timeEnd.Sub(timeStart), nil
}

func durationOneWorkday() time.Duration {
	workHoursPerWeek, err := time.ParseDuration(viper.GetString("worktime_per_week"))
	if err != nil {
		log.Println(fmt.Errorf("workHoursMonth: %w", err))
	}
	workHoursPerDay := workHoursPerWeek / 5
	return workHoursPerDay
}

func findIDProject(id int) int {
	for index, project := range projectList {
		if project.ID == id {
			return index
		}
	}
	return -1
}

func workTotalByDate(date time.Time) time.Duration {
	timeTotal, err := time.ParseDuration("0s")
	if err != nil {
		log.Print(err)
	}
	records := database.GetRecordsForDate(date)
	for _, record := range records {
		timeStart, err := time.Parse("15:04", record.Start)
		if err != nil {
			log.Print(err)
		}
		timeEnd, err := time.Parse("15:04", record.End)
		if err != nil {
			log.Print(err)
		}
		diffTime := timeEnd.Sub(timeStart)
		timeTotal += diffTime
	}
	return timeTotal
}

func workTotalForRecords(timeframes []database.Timeframe) time.Duration {
	timeTotal, err := time.ParseDuration("0s")
	if err != nil {
		log.Print(err)
	}
	for _, record := range timeframes {
		timeStart, err := time.Parse("15:04", record.Start)
		if err != nil {
			log.Print(err)
		}
		timeEnd, err := time.Parse("15:04", record.End)
		if err != nil {
			log.Print(err)
		}
		diffTime := timeEnd.Sub(timeStart)
		timeTotal += diffTime
	}
	return timeTotal
}

func weekDaysByDate(date time.Time) []time.Time {
	var weekDays []time.Time
	daysInAWeek := []int{1, 2, 3, 4, 5, 6, 7}

	for _, weekday := range daysInAWeek {
		dayOfWeek := int(date.Weekday())
		if dayOfWeek == 0 {
			dayOfWeek = 7
		}
		weekdayOffset := weekday - dayOfWeek
		weekDays = append(weekDays, date.AddDate(0, 0, weekdayOffset))
	}
	return weekDays
}

func weekWorkdaysByDate(date time.Time) float64 {
	var workDays float64
	weekDays := weekDaysByDate(date)

	for _, day := range weekDays {
		if calendar.IsWorkday(day) {
			workDays++
		}
	}
	return workDays
}

func workTotalWeek(date time.Time) time.Duration {
	workTotalDuration, err := time.ParseDuration("0s")
	if err != nil {
		log.Print(err)
	}
	var daysInWeek []time.Time = weekDaysByDate(date)
	for _, day := range daysInWeek {
		dayDuration := workTotalByDate(day)
		workTotalDuration += dayDuration
	}
	return workTotalDuration
}

func workDeltaWeek(workTotalDuration time.Duration, date time.Time) time.Duration {
	workTotalTarget, err := time.ParseDuration(viper.GetString("worktime_per_week"))
	if err != nil {
		log.Println(err)
	}
	workTotalTarget, err = time.ParseDuration(fmt.Sprintf("%fh", workTotalTarget.Hours()/5.0*weekWorkdaysByDate(date)))
	workDelta := workTotalDuration - workTotalTarget
	if err != nil {
		log.Println(err)
	}
	return workDelta
}

func convertTimeframesForPlanner(timeframes []database.Timeframe) []database.PlannerEntry {
	start, _ := time.Parse("2006-01-02", fmt.Sprintf("%d-01-01", activeYearSummary.Year()))
	end, _ := time.Parse("2006", fmt.Sprintf("%d", activeYearSummary.Year()+1))
	dayTime := time.Hour * 24
	dateStringsWhole := []string{}
	dateStringsHalf := []string{}
	returnArray := []database.PlannerEntry{}

	for _, tf := range timeframes {
		diffTime, err := timeframeDuration(tf)
		if err != nil {
			log.Println(err)
		}

		if diffTime < durationOneWorkday() {
			dateStringsHalf = append(dateStringsHalf, tf.Date)
		} else {
			dateStringsWhole = append(dateStringsWhole, tf.Date)
		}
	}

	for cur := start; cur.Before(end); cur = cur.Add(dayTime) {
		isHoliday, _, _ := calendar.IsHoliday(cur)
		if slices.Contains(dateStringsWhole, cur.Format("2006-01-02")) {
			returnArray = append(returnArray, database.PlannerEntry{Date: cur.Format("2006-01-02"), DayType: 1})
		} else if slices.Contains(dateStringsHalf, cur.Format("2006-01-02")) {
			returnArray = append(returnArray, database.PlannerEntry{Date: cur.Format("2006-01-02"), DayType: 2})
		} else if isHoliday {
			returnArray = append(returnArray, database.PlannerEntry{Date: cur.Format("2006-01-02"), DayType: 3})
		} else if cal.IsWeekend(cur) {
			returnArray = append(returnArray, database.PlannerEntry{Date: cur.Format("2006-01-02"), DayType: 4})
		} else {
			returnArray = append(returnArray, database.PlannerEntry{Date: cur.Format("2006-01-02"), DayType: 0})
		}

	}
	return returnArray
}

// Returns duration in hours for all project in month, including base projects
func GetProjectHoursMonth(month time.Time) []database.ProjectHours {
	projectsList := []database.ProjectHours{}
	projects := database.GetProjects()
	total, err := time.ParseDuration("0s")
	if err != nil {
		log.Println(err)
	}
	for _, project := range projects {
		records := database.GetRecordsForProjectAndMonth(month.Year(), int(month.Month()), project.ID)
		duration := workTotalForRecords(records)
		total += duration
		projectHour := database.ProjectHours{
			Project: project,
			Hours:   fmt.Sprintf("%.2f", duration.Hours()),
		}
		if projectHour.ID < 5 || duration.Hours() > 0 {
			projectsList = append(projectsList, projectHour)
		}
	}
	projectsList = append(projectsList, database.ProjectHours{
		Project: database.Project{Activity: "", Details: "", Name: "Total"},
		Hours:   fmt.Sprintf("%.2f", total.Hours()),
	})
	return projectsList
}

// Returns duration in hours for all custom projects in year, excluding base projects
func GetProjectHoursYear(year time.Time) []database.ProjectHours {
	projectsList := []database.ProjectHours{}
	projects := database.GetProjects()
	total, err := time.ParseDuration("0s")
	if err != nil {
		log.Println(err)
	}
	for _, project := range projects {
		records := database.GetRecordsForProjectAndYear(year, project.ID)
		duration := workTotalForRecords(records)
		total += duration
		projectHour := database.ProjectHours{
			Project: project,
			Hours:   fmt.Sprintf("%.2f", duration.Hours()),
		}
		if project.ID < 5 || duration.Hours() > 0 {
			projectsList = append(projectsList, projectHour)
		}
	}
	projectsList = append(projectsList, database.ProjectHours{
		Project: database.Project{Activity: "", Details: "", Name: "Total"},
		Hours:   fmt.Sprintf("%.2f", total.Hours()),
	})
	return projectsList
}

func GetProjectsHoursOverview(month time.Time) []database.ProjectHoursLocationsDaily {
	var dailyEntries []database.ProjectHoursLocationsDaily
	lastDayMonth := cal.MonthEnd(month)
	days := lastDayMonth.Day()
	for day := range days {
		dayTime, err := time.Parse("2006-01-02", fmt.Sprintf("%s-%02d", month.Format("2006-01"), day+1))
		if err != nil {
			log.Println(err)
		}
		workTotal := workTotalByDate(dayTime)
		projectEntries := database.GetProjectsForDate(dayTime)
		projectStrings := []string{}
		for _, entry := range projectEntries {
			projectStrings = append(projectStrings, entry)
		}
		locationEntries := database.GetLocationsForDate(dayTime)
		locationStrings := []string{}
		for _, entry := range locationEntries {
			locationStrings = append(locationStrings, entry)
		}

		dailyEntries = append(dailyEntries,
			database.ProjectHoursLocationsDaily{
				Date:      dayTime.Format("Mon Jan 2 2006"),
				Hours:     fmt.Sprintf("%05.2f", workTotal.Hours()),
				Projects:  projectStrings,
				Locations: locationStrings})

	}

	return dailyEntries
}

func GetOvertimeHoursUntilDay(year time.Time, day time.Time) database.ProjectHours {
	total, err := time.ParseDuration("0s")
	if err != nil {
		log.Println(err)
	}

	records := database.GetRecordsUntilDay(day)
	duration := workTotalForRecords(records)
	total += duration

	workTotalTarget, err := time.ParseDuration(viper.GetString("worktime_per_week"))
	if err != nil {
		log.Println(err)
	}

	// TODO: Write Work Hours per Day into Records? Extra days table?
	offsetOvertime, err := time.ParseDuration(viper.GetString("offset_overtime"))
	if err != nil {
		log.Println(err)
	}

	// Calculate work days
	years := database.GetYears()
	workdays := 0
	for _, yearDB := range years {
		if yearDB <= year.Year() {
			year_as_time, _ := time.Parse("2006", fmt.Sprintf("%d", yearDB))
			workdays += GetWorkDaysForYearUntilDay(year_as_time, day)
		}
	}

	overTimeHours := offsetOvertime.Hours() + total.Hours() - float64(workdays)*workTotalTarget.Hours()/5
	return database.ProjectHours{
		Project: database.Project{Activity: "", Details: "", Name: "Overtime"},
		Hours:   fmt.Sprintf("%.2f", overTimeHours),
	}
}

func GetWorkDays(month time.Time) int {
	day, _ := time.Parse("2006-01-02", month.Format("2006-01")+"-01")
	workdays := calendar.WorkdaysRemain(day)
	if calendar.IsWorkday(day) {
		return workdays + 1
	} else {
		return workdays
	}
}

// If year is smaller than year of day, work days for whole year are returned
func GetWorkDaysForYearUntilDay(year time.Time, day time.Time) int {
	startOfYear, err := time.Parse("2006-01-02", fmt.Sprintf("%d-01-01", year.Year()))
	if err != nil {
		log.Println(err)
	}
	var endDate time.Time
	if year.Year() < day.Year() {
		endDate, err = time.Parse("2006-01-02", fmt.Sprintf("%d-12-31", year.Year()))
		if err != nil {
			log.Println(err)
		}
	} else {
		dayParse, err := time.Parse("2006-01-02", day.Format("2006-01-02"))
		if err != nil {
			log.Printf("error parse day %v", err)
		}
		endDate = dayParse
	}
	return calendar.WorkdaysInRange(startOfYear, endDate)
}
