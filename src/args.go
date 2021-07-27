package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sqweek/dialog"
)

var usageTemplate = `
Neoray is an ui client for neovim.
Copyright (c) 2021 Ismail Bulut.
Version %d.%d.%d %s
License %s
Webpage %s

Options:

--file <name>
	Filename to open.
--line <number>
	Goto line number.
--column <number>
	Goto column number.
--singleinstance, -si
	Only accept one instance of neoray and send all flags to it.
--verbose <filename>
	Specify a filename to verbose debug output.
--nvim <path>
	Path to nvim executable. May be relative or absolute.
--help, -h
	Prints this message and quits.

All other flags will send to neovim.
`

type ParsedArgs struct {
	file       string
	line       int
	column     int
	singleinst bool
	execPath   string
	others     []string
}

func ParseArgs(args []string) ParsedArgs {
	// Init defaults
	options := ParsedArgs{
		file:       "",
		line:       -1,
		column:     -1,
		singleinst: false,
		execPath:   "nvim",
		others:     []string{},
	}
	var err error
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--file":
			assert(len(args) > i+1, "specify filename after --file")
			options.file = args[i+1]
			i++
			break
		case "--line":
			assert(len(args) > i+1, "specify line number after --line")
			options.line, err = strconv.Atoi(args[i+1])
			assert(err == nil, "invalid number after --line")
			i++
			break
		case "--column":
			assert(len(args) > i+1, "specify column number after --column")
			options.column, err = strconv.Atoi(args[i+1])
			assert(err == nil, "invalid number after --column")
			i++
			break
		case "--singleinstance", "-si":
			options.singleinst = true
			break
		case "--nvim":
			assert(len(args) > i+1, "specify path after --nvim")
			absolute, err := filepath.Abs(args[i+1])
			if err == nil {
				options.execPath = absolute
			}
			i++
			break
		case "--verbose":
			assert(len(args) > i+1, "specify filename after --verbose")
			initVerboseFile(args[i+1])
			i++
			break
		case "--help", "-h":
			PrintHelp()
			os.Exit(0)
			break
		default:
			options.others = append(options.others, args[i])
			break
		}
	}
	return options
}

func PrintHelp() {
	// About
	msg := fmt.Sprintf(usageTemplate,
		VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH,
		buildTypeString(), LICENSE, WEBPAGE)
	dialog.Message(msg).Title("Neoray Help").Info()
}

// Call this before starting neovim.
func (options ParsedArgs) ProcessBefore() bool {
	if options.singleinst {
		// First we will check only once because sending and
		// waiting http requests will make neoray opens slower.
		client, err := CreateClient()
		if err != nil {
			logDebug("No instance found or tcp client creation failed:", err)
			return false
		}
		ok := false
		if client.sendSignal(SIGNAL_CHECK_CONNECTION) {
			if options.file != "" {
				fullPath, err := filepath.Abs(options.file)
				if err == nil {
					client.sendSignal(SIGNAL_OPEN_FILE, fullPath)
				}
			}
			if options.line != -1 {
				client.sendSignal(SIGNAL_GOTO_LINE, strconv.Itoa(options.line))
			}
			if options.column != -1 {
				client.sendSignal(SIGNAL_GOTO_COLUMN, strconv.Itoa(options.column))
			}
			ok = true
		}
		client.Close()
		return ok
	}
	return false
}

// Call this after connected neovim as ui.
func (options ParsedArgs) ProcessAfter() {
	if options.singleinst {
		server, err := CreateServer()
		if err != nil {
			logMessage(LOG_LEVEL_ERROR, LOG_TYPE_NEORAY, "Failed to create TCP listener:", err)
		} else {
			singleton.server = server
		}
	}
	if options.file != "" {
		singleton.nvim.openFile(options.file)
	}
	if options.line != -1 {
		singleton.nvim.gotoLine(options.line)
	}
	if options.column != -1 {
		singleton.nvim.gotoColumn(options.column)
	}
}
