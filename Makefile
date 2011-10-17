include $(GOROOT)/src/Make.inc

TARG=scouting
GOFILES=\
	main.go\
	model.go\
	team.go\

include $(GOROOT)/src/Make.cmd
