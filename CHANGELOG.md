# Changelog

## v0.5.7

### Features

- Deleted projects are only hidden from projects page through is_deleted flag
- First optimization for mobile design
- Included weekday in create record card

### Bugfix

- Default decimal separator change to ',' to write a proper default config

## v0.5.6

### Bugfix

- Superfluous HTTP header error
- fixed overtime calculation across years

---

## v0.5.5

### Features

- Export button for Monthly Summary
- Support for Linux

---

## v0.5.4

### Features

- Added arrows to change to next/previous day in create record card

### Bugfix

- Fixed further database lock-ups

---

## v0.5.3

### Features

- disabled enter button for create records form
- improved visuals for daily overview on month page
- frontend now handles time formatting
- removed seconds from daily summary

---

## v0.5.2

### Features

- decimal_separator: "." in timetracker.yaml can be used to change separator in copy clipboard function

### Bugfix

- database constraints updated

---

## v0.5.0

### Features

- Vacation shortcut button in records

### Bugfix

- Improved error handling

---

## v0.4.1

### Features

- Version String in Headline
- Total Work Delta in Daily Summary
- Logging to file can be enabled for debugging purpose (timetracker.yaml: "logfile: true")

### Bugfix

- Fixed crash at startup due to database lock

---
