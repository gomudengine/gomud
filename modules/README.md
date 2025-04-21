# Modules

Extract any modules into this folder.

* Modules should be named uniquely, in a manner that identifies their purpose.
* Modules should be inside of a subfolder of `modules`, named after their package.
  * Example: `modules/birds/` would contain the `birds` module/package.

* Module folders should container a `datafiles` folder that contains any datafiles needed.
  * Files within `datafiles` will be treated as though located within the actual `_datafiles`
  * These files are read-only.

## Things Modules can do:

* Provide template files
* Add or Override user commands or mob commands
* Save/Load their own data
* Track their own config values
* Modify help menu
* Add help aliases
* Add command aliases
* Listen for/Handle events
* Access the rest of the code

# Examples

## Basic user command function

* time/time.go
* time/files/*

## User command with maintained state and save/loading of data

* leaderboards/leaderboards.go
* leaderboards/files/*
